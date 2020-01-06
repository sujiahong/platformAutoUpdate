
import ftplib,os

def ftpDirCreate(ftp, filepath):
    lst = filepath.split("/")
    pth = ""
    for name in lst:
        pth += name + "/"
        print(pth)
        try:
            ftp.cwd(pth)
        except Exception as e:
            print("-===--=err:   ", e)
            ftp.mkd(pth)

def ftpDirCreate1(ftp, filepath):
    lst = filepath.split("/")
    pth = ""
    for name in lst:
        pth += name + "/"
        if name == '':
            continue
        try:
            if name in ftp.nlst():
                ftp.cwd(pth)
            else:
                ftp.mkd(pth)
                ftp.cwd(pth)
        except Exception as e:
            print("err:   ", e)

def up():
    try:
        ftp = ftplib.FTP()
        ftp.connect("10.0.1.18", 21, 5)
        ftp.login("winftp", "Gloud@123")
        # ftp.cwd("/u/0")
        # print(ftp.pwd(), ftp.dir())
        # ftp.mkd("/u")
        ftpDirCreate1(ftp, "/u/0")
        ftp.cwd("/u/0")
        upfp = open("E:\\rsync-3.1.3\\lib\\compat.o", "rb")
        ftp.storbinary("STOR "+ "compat.o", upfp)
        upfp.close()
        ftp.quit()
    except Exception as e:
        print("up up up   ", e)

# up()

def down():
    try:
        ftp = ftplib.FTP()
        ftp.connect("10.0.1.18", 21, 5)
        ftp.login("winftp", "Gloud@123")
        ftp.cwd("/u/0")
        lst = ftp.nlst()
        print(lst, len(lst))
        downfp = open("E:\\compat.o", "wb")
        ftp.retrbinary("RETR "+"compat.o", downfp.write)
        downfp.close()
        ftp.quit()
    except Exception as e:
        print("down down   ", e)

# down()

ret = os.system('\"C:\Program Files\Rockstar Games\Launcher\Redistributables\SocialClub\Social-Club-Setup.exe\" /silent')
print(ret)
