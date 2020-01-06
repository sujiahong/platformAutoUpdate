package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	//"github.com/StackExchange/wmi"
	"github.com/Unknwon/goconfig"
)

var configMap = make(map[string]map[string]string)
var blackList []string

func isFileExist(path string) bool {
	finfo, err := os.Stat(path)
	fmt.Println(finfo, err)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func isInBlackList(lst []string, s string) bool {
	for _, name := range lst {
		if s == name {
			return true
		}
	}
	return false
}

func clearSteamOldFile() error {
	finfoArr, err := ioutil.ReadDir(configMap["steam"]["src_folder"])
	if err != nil {
		fmt.Println("clearFile error: ", err)
		return err
	}
	//var blackList = strings.Split(configMap["updater"]["black_type"], ",")
	for _, info := range finfoArr {
		if !info.IsDir() {
			if isInBlackList(blackList, filepath.Ext(info.Name())) {
				err = os.Remove(configMap["steam"]["src_folder"] + "\\" + info.Name())
				fmt.Println("remove err: ", err)
			}
		}
	}
	return nil
}

func walkCallback(path string, info os.FileInfo, err error) error {
	var retainFile = configMap["steam"]["pakcage_retain"]
	var fileList = strings.Split(retainFile, ",")
	if info.Name() == fileList[0] || info.Name() == fileList[1] {
		return err
	}
	os.Remove(path)
	return err
}

func clearSteamPackage() {
	var err = filepath.Walk(configMap["steam"]["package_folder"], walkCallback)
	fmt.Println("clearSteamPackage err === ", err)
}

func runPlatform(pf string) {
	fmt.Println("run ", pf)
	cmd := exec.Command(configMap[pf]["check_file"])
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("run platform err: ", err, errbuf.String())
		return
	}
	fmt.Println("run 平台成功!")
}

func isSteamUpdate() bool {
	finfoArr, err := ioutil.ReadDir(configMap["steam"]["package_folder"])
	fmt.Println(finfoArr, err, len(finfoArr))
	if len(finfoArr) > 2 {
		var retainFile = configMap["steam"]["pakcage_retain"]
		var fileList = strings.Split(retainFile, ",")
		for _, info := range finfoArr {
			fmt.Println(info.Name())
			if info.Name() != fileList[0] && info.Name() != fileList[1] {
				return true
			}
		}
	}
	return false
}

func getSteamVersion2() string {
	var cmd = exec.Command("cmd", "/C", "wmic datafile where `name=D:\\soft\\Steam\\Steam.exe` get version")
	var errbuf, outbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err: ", err, errbuf.String())
		return ""
	}
	fmt.Println("7777777   ", outbuf.String())
	return "0"
}

func utf16ToString(b []byte, bom int) string {
	if len(b) >= 2 {
		switch n := uint16(b[0])<<8 | uint16(b[1]); n {
		case 0xfffe:
			fallthrough
		case 0xfeff:
			b = b[2:]
			break
		default:
			b = b[1:]
		}
	}
	utf16Arr := make([]uint16, len(b)/2)
	for i := range utf16Arr {
		utf16Arr[i] = uint16(b[2*i+bom&1])<<8 | uint16(b[2*i+(bom+1)&1])
	}
	return string(utf16.Decode(utf16Arr))
}

func getSteamVersion() string {
	var cmd = exec.Command(".\\wmicVersion.bat")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err: ", err, errbuf.String())
		return ""
	}
	fp, err := os.Open("./version.txt")
	if err != nil {
		fmt.Println("open file error: ", err)
		return ""
	}
	defer fp.Close()
	br := bufio.NewReader(fp)
	for {
		byteArr, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		str := utf16ToString(byteArr, 1)
		idx := strings.Index(str, ".")
		if idx > -1 {
			return str
		}
	}
	return ""
}

func closeSteam() {
	fmt.Println("close steam")
	var cmd = exec.Command(configMap["steam"]["check_file"], "-shutdown")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err: ", err, errbuf.String())
		return
	}
	fmt.Println("close steam 关闭成功")
}

func doCompress(scn, f string) error {
	cmd := exec.Command(".\\bin\\7z.exe", "a", configMap[scn]["zip_name"], configMap[scn]["src_folder"]+f)
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err: ", err, errbuf.String())
		return err
	}
	return nil
}

func compressSteam() {
	fmt.Println("compress steam")
	finfoArr, err := ioutil.ReadDir(configMap["steam"]["src_folder"])
	if err != nil {
		fmt.Println("clearFile error: ", err)
		return
	}
	var steamBlackList = strings.Split(configMap["steam"]["black_list"], ",")
	for _, info := range finfoArr {
		//fmt.Println("-----   ", info.Name())
		if !isInBlackList(steamBlackList, info.Name()) {
			err = doCompress("steam", info.Name())
			if err != nil {
				fmt.Println("打包压缩失败停止压缩！")
				break
			}
		}
	}
}

func getFilemd5(f string) string {
	h := md5.New()
	fd, err := os.Open(f)
	if err != nil {
		fmt.Println("打开文件错误  ", err)
		return ""
	}
	defer fd.Close()
	// var readbuf = make([]byte, 4096)
	// for {
	// 	n, err := fd.Read(readbuf)
	// 	if err != nil {
	// 		fmt.Println("读文件 buf error: ", err)
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		return ""
	// 	}
	// 	fmt.Println("n= ", n)
	// 	h.Write(readbuf[:n])
	// }
	io.Copy(h, fd)
	str := fmt.Sprintf("%x", h.Sum(nil))
	fmt.Println("fisufjs     ", str)
	return str
}

func writeConfigAndSteamCompress(version string) {
	preVersion := configMap["steam"]["check_type"]
	if preVersion != version {
		configMap["steam"]["check_type"] = version
		cfger, err := goconfig.LoadConfigFile("./config.ini")
		cfger.SetValue("steam", "check_type", version)
		goconfig.SaveConfigFile(cfger, "./config.ini")
		compressSteam()
		md5str := getFilemd5("./" + configMap["steam"]["zip_name"])
		cfg, err := goconfig.LoadConfigFile("./gloud_game_platforms.ini")
		if err != nil {
			fmt.Println("读取ini配置文件错误！！！")
			return
		}
		cfg.SetValue("Steam_online", "hashfile_md5", md5str)
		ver, err := cfg.GetValue("Steam_online", "version")
		if err != nil {
			fmt.Println("获取gloud_game_platforms.ini, Steam_online, version值错误")
			return
		}
		iver, err := strconv.Atoi(ver)
		if err != nil {
			fmt.Println("字符串转换成整型出错！！！")
			return
		}
		iver = iver + 1
		cfg.SetValue("Steam_online", "version", strconv.Itoa(iver))
		ver, err = cfg.GetValue("Steam_offline", "version")
		if err != nil {
			fmt.Println("获取gloud_game_platforms.ini, Steam_offline, version值错误")
			return
		}
		iver, err = strconv.Atoi(ver)
		if err != nil {
			fmt.Println("字符串转换成整型出错！！！")
			return
		}
		iver = iver + 1
		cfg.SetValue("Steam_offline", "version", strconv.Itoa(iver))
		goconfig.SaveConfigFile(cfg, "./gloud_game_platforms.ini")
	} else {
		fmt.Println("没有更新 ")
	}
}

func checkSteamUpdate() {
	clearSteamOldFile()
	clearSteamPackage()
	go runPlatform("steam")
	time.Sleep(600 * time.Second)
	if isSteamUpdate() {
		closeSteam()
		var version = getSteamVersion()
		fmt.Println("99999  ", version)
		writeConfigAndSteamCompress(version)
	} else {
		fmt.Println("steam 没有更新")
		closeSteam()
	}
}

/////////////////////////////////////////////////////////////

func clearRockstarOldFile() error {
	var srcList = strings.Split(configMap["RockStarx64"]["src_list"], ",")
	for _, ph := range srcList {
		finfoArr, err := ioutil.ReadDir(ph)
		if err != nil {
			fmt.Println("clearFile error: ", err)
			return err
		}
		//var blackList = strings.Split(configMap["updater"]["black_type"], ",")
		for _, info := range finfoArr {
			//fmt.Println(info.Name())
			if !info.IsDir() {
				if isInBlackList(blackList, filepath.Ext(info.Name())) {
					err = os.Remove(configMap["RockStarx64"]["src_list"] + "\\" + info.Name())
					fmt.Println("clearRockstarOldFile remove err: ", err)
				}
			}
		}
	}
	return nil
}

func isRockstarUpdate() bool {
	var srcList = strings.Split(configMap["RockStarx64"]["src_folder"], ",")
	for _, ph := range srcList {
		finfoArr, err := ioutil.ReadDir(ph)
		if err != nil {
			fmt.Println("clearFile error: ", err)
			return false
		}
		//var blackList = strings.Split(configMap["updater"]["black_type"], ",")
		for _, info := range finfoArr {
			fmt.Println(info.Name())
			if !info.IsDir() {
				if isInBlackList(blackList, filepath.Ext(info.Name())) {
					return true
				}
			}
		}
	}
	return false
}

func closeRockstar() {
	fmt.Println("close rockstar")
	var cmd = exec.Command("taskkill", "/f", "/im", "LauncherPatcher.exe")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("closeRockstar LauncherPatcher.exe err: ", err, errbuf.String())
		return
	}
	cmd = exec.Command("taskkill", "/f", "/im", "Launcher.exe")
	cmd.Stderr = &errbuf
	err = cmd.Run()
	if err != nil {
		fmt.Println("closeRockstar Launcher.exe err: ", err, errbuf.String())
		return
	}
	fmt.Println("close rockstar 关闭成功")
}

func getRockstarVersion() string {
	var cmd = exec.Command(".\\rockstarWMICVersion.bat")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err: ", err, errbuf.String())
		return ""
	}
	fp, err := os.Open("./rockstar_version.txt")
	if err != nil {
		fmt.Println("open file error: ", err)
		return ""
	}
	defer fp.Close()
	br := bufio.NewReader(fp)
	for {
		byteArr, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		str := utf16ToString(byteArr, 1)
		idx := strings.Index(str, ".")
		if idx > -1 {
			return str
		}
	}
	return ""
}

func rockstarAdditionCmd() {
	fmt.Println("rockstarAdditionCmd")
	var cmd = exec.Command(configMap["RockStarx64"]["addition_cmd"])
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("rockstarAdditionCmd err: ", err, errbuf.String())
		return
	}
}

func compressRockstar() {
	fmt.Println("compress rockstar")
	finfoArr, err := ioutil.ReadDir(configMap["RockStarx64"]["src_folder"])
	if err != nil {
		fmt.Println("clearFile error: ", err)
		return
	}
	for _, info := range finfoArr {
		fmt.Println("-----   ", info.Name())
		err = doCompress("RockStarx64", info.Name())
		if err != nil {
			fmt.Println("打包压缩失败停止压缩！")
			break
		}
	}
	finfoArrx86, err := ioutil.ReadDir(configMap["RockStarx86"]["src_folder"])
	if err != nil {
		fmt.Println("clearFile error: ", err)
		return
	}
	for _, info := range finfoArrx86 {
		fmt.Println("-----   ", info.Name())
		err = doCompress("RockStarx86", info.Name())
		if err != nil {
			fmt.Println("打包压缩失败停止压缩！")
			break
		}
	}
}

func writeConfigAndRockstarCompress(version string) {
	preVersion := configMap["RockStarx64"]["check_type"]
	if preVersion != version {
		configMap["RockStarx64"]["check_type"] = version
		cfger, err := goconfig.LoadConfigFile("./config.ini")
		cfger.SetValue("RockStarx64", "check_type", version)
		goconfig.SaveConfigFile(cfger, "./config.ini")
		rockstarAdditionCmd()
		compressRockstar()
		md5str := getFilemd5("./" + configMap["RockStarx64"]["zip_name"])
		cfg, err := goconfig.LoadConfigFile("./gloud_game_platforms.ini")
		if err != nil {
			fmt.Println("读取ini配置文件错误！！！")
			return
		}
		cfg.SetValue("rockstar-x64", "hashfile_md5", md5str)
		md5strx86 := getFilemd5("./" + configMap["RockStarx86"]["zip_name"])
		cfg.SetValue("rockstar-x86", "hashfile_md5", md5strx86)

		ver, err := cfg.GetValue("rockstar-x64", "version")
		if err != nil {
			fmt.Println("获取gloud_game_platforms.ini, rockstar-x64, version值错误")
			return
		}
		iver, err := strconv.Atoi(ver)
		if err != nil {
			fmt.Println("字符串转换成整型出错！！！")
			return
		}
		iver = iver + 1
		cfg.SetValue("rockstar-x64", "version", strconv.Itoa(iver))
		ver, err = cfg.GetValue("rockstar-x86", "version")
		if err != nil {
			fmt.Println("获取gloud_game_platforms.ini, rockstar-x86, version值错误")
			return
		}
		iver, err = strconv.Atoi(ver)
		if err != nil {
			fmt.Println("字符串转换成整型出错！！！")
			return
		}
		iver = iver + 1
		cfg.SetValue("rockstar-x86", "version", strconv.Itoa(iver))
		goconfig.SaveConfigFile(cfg, "./gloud_game_platforms.ini")
	} else {
		fmt.Println("没有更新 ")
	}
}

func checkRockstarX64Update() {
	clearRockstarOldFile()
	go runPlatform("RockStarx64")
	time.Sleep(600 * time.Second)
	if isRockstarUpdate() {
		closeRockstar()
		var version = getRockstarVersion()
		fmt.Println("99999  ", version)
		writeConfigAndRockstarCompress(version)
	} else {
		fmt.Println("rockstar 没有更新")
		closeRockstar()
	}
	// var version = getRockstarVersion()
	// fmt.Println("99999  ", version)
	// writeConfigAndRockstarCompress(version)
}

func checkRockstarX86Update() {
}

func clearOldFile(ph string) {
	finfoArr, err := ioutil.ReadDir(ph)
	if err != nil {
		fmt.Println("clearFile error: ", err)
		return
	}
	for _, info := range finfoArr {
		//fmt.Println(ph, info.Name())
		if !info.IsDir() {
			if isInBlackList(blackList, filepath.Ext(info.Name())) {
				err = os.Remove(ph + info.Name())
				fmt.Println("clearOldFile remove err: ", err)
			}
		} else {
			clearOldFile(ph + info.Name() + "\\")
		}
	}
}

func isPlatformUpdate(ph string) bool {
	finfoArr, err := ioutil.ReadDir(ph)
	if err != nil {
		fmt.Println("isPlatformUpdate error: ", err)
		return false
	}
	for _, info := range finfoArr {
		//fmt.Println(ph, info.Name())
		if !info.IsDir() {
			if isInBlackList(blackList, filepath.Ext(info.Name())) {
				return true
			}
		} else {
			clearOldFile(ph + info.Name() + "\\")
		}
	}
	return false
}

func closeEpic() {
	fmt.Println("close epic")
	var cmd = exec.Command("taskkill", "/f", "/im", "EpicGamesLauncher.exe")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("closeEpic EpicGamesLauncher.exe err: ", err, errbuf.String())
		return
	}
	fmt.Println("close epic 关闭成功")
}

func getEpicVersion() string {
	var cmd = exec.Command(".\\epicWMICVersion.bat")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err: ", err, errbuf.String())
		return ""
	}
	fp, err := os.Open("./epic_version.txt")
	if err != nil {
		fmt.Println("open file error: ", err)
		return ""
	}
	defer fp.Close()
	br := bufio.NewReader(fp)
	for {
		byteArr, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		str := utf16ToString(byteArr, 1)
		idx := strings.Index(str, ".")
		if idx > -1 {
			return str
		}
	}
	return ""
}

func compressPlatform(section string) {
	fmt.Println("compress platform")
	finfoArr, err := ioutil.ReadDir(configMap[section]["src_folder"])
	if err != nil {
		fmt.Println("clearFile error: ", err)
		return
	}
	for _, info := range finfoArr {
		fmt.Println("-----   ", info.Name())
		err = doCompress(section, info.Name())
		if err != nil {
			fmt.Println("打包压缩失败停止压缩！")
			break
		}
	}
}

func rewriteConfigAndCompress(version, section string) {
	preVersion := configMap[section]["check_type"]
	if preVersion != version {
		configMap[section]["check_type"] = version
		cfger, err := goconfig.LoadConfigFile("./config.ini")
		cfger.SetValue(section, "check_type", version)
		goconfig.SaveConfigFile(cfger, "./config.ini")
		compressPlatform(section)
		md5str := getFilemd5("./" + configMap[section]["zip_name"])
		cfg, err := goconfig.LoadConfigFile("./gloud_game_platforms.ini")
		if err != nil {
			fmt.Println("读取ini配置文件错误！！！")
			return
		}
		cfg.SetValue(section, "hashfile_md5", md5str)

		ver, err := cfg.GetValue(section, "version")
		if err != nil {
			fmt.Println("获取gloud_game_platforms.ini, version值错误")
			return
		}
		iver, err := strconv.Atoi(ver)
		if err != nil {
			fmt.Println("字符串转换成整型出错！！！")
			return
		}
		iver = iver + 1
		cfg.SetValue(section, "version", strconv.Itoa(iver))
		goconfig.SaveConfigFile(cfg, "./gloud_game_platforms.ini")
	} else {
		fmt.Println("没有更新 ")
	}
}

func checkEpicUpdate() {
	clearOldFile(configMap["Epic"]["src_folder"])
	go runPlatform("Epic")
	time.Sleep(600 * time.Second)
	if isPlatformUpdate(configMap["Epic"]["src_folder"]) {
		closeEpic()
		var version = getEpicVersion()
		fmt.Println("99999  ", version)
		rewriteConfigAndCompress(version, "Epic")
	} else {
		fmt.Println("epic 没有更新")
		closeEpic()
	}
	var version = getEpicVersion()
	fmt.Println("99999  ", version)
	rewriteConfigAndCompress(version, "Epic")
}

func closeUbisoft() {
	fmt.Println("close ubisoft")
	var cmd = exec.Command("taskkill", "/f", "/im", "upc.exe")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("closeUbisoft upc.exe err: ", err, errbuf.String())
		return
	}
	fmt.Println("close ubisoft 关闭成功")
}

func getUbisoftVersion() string {
	var cmd = exec.Command(".\\ubisoftWMICVersion.bat")
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err: ", err, errbuf.String())
		return ""
	}
	fp, err := os.Open("./ubisoft_version.txt")
	if err != nil {
		fmt.Println("open file error: ", err)
		return ""
	}
	defer fp.Close()
	br := bufio.NewReader(fp)
	for {
		byteArr, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		str := utf16ToString(byteArr, 1)
		idx := strings.Index(str, ".")
		if idx > -1 {
			return str
		}
	}
	return ""
}

func checkUbisoftUpdate() {
	clearOldFile(configMap["Uplay"]["src_folder"])
	go runPlatform("Uplay")
	time.Sleep(600 * time.Second)
	if isPlatformUpdate(configMap["Uplay"]["src_folder"]) {
		closeUbisoft()
		var version = getUbisoftVersion()
		fmt.Println("99999  ", version)
		rewriteConfigAndCompress(version, "Uplay")
	} else {
		fmt.Println("Uplay 没有更新")
		closeUbisoft()
	}
	var version = getUbisoftVersion()
	fmt.Println("99999  ", version)
	rewriteConfigAndCompress(version, "Uplay")
}

func readINI() error {
	cfg, err := goconfig.LoadConfigFile("./config.ini")
	if err != nil {
		return err
	}
	glob, err := cfg.GetSection("updater")
	//fmt.Println(glob, err)
	configMap["updater"] = glob
	glob, err = cfg.GetSection("RockStarx64")
	configMap["RockStarx64"] = glob
	glob, err = cfg.GetSection("RockStarx86")
	configMap["RockStarx86"] = glob
	glob, err = cfg.GetSection("Epic")
	configMap["Epic"] = glob
	glob, err = cfg.GetSection("Uplay")
	configMap["Uplay"] = glob
	glob, err = cfg.GetSection("steam")
	configMap["steam"] = glob
	fmt.Println(configMap["steam"]["zip_name"])
	return err
}

func main() {
	fmt.Println("开始更新！！！")
	readINI()
	blackList = strings.Split(configMap["updater"]["black_type"], ",")
	for {
		checkSteamUpdate()
		checkRockstarX64Update()
		checkEpicUpdate()
		checkUbisoftUpdate()
		time.Sleep(10 * time.Second)
	}
}
