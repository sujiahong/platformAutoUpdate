import subprocess,os,time,configparser,re,threading,logging,ctypes,platform


class VersionType(str):
    def __new__(cls, *args, **kwargs):
        return super().__new__(cls)
    def __init__(self,string:str):
        str.__init__(self)
        try:
            self.version_list = [ int(i) for i in string.split('.') if i.isalnum()]
        except Exception as e :
            print(e)
            self.version_list = [0]
        self.lenth = len(self.version_list)
    def __lt__(self,other):
        for index in range(self.lenth):
            if self.version_list[index] < other.version_list[index]:
                return True
        return False
    def __eq__(self,other):
        for index in range(self.lenth):
            if self.version_list[index] != other.version_list[index]:
                return False
        return True
    def __gt__(self,other):
        for index in range(self.lenth):
            if self.version_list[index] > other.version_list[index]:
                return True
        return False
    def __le__(self,other):
        for index in range(self.lenth):
            if self.version_list[index] > other.version_list[index]:
                return False
        return True
    def __ne__(self,other):
        for index in range(self.lenth):
            if self.version_list[index] != other.version_list[index]:
                return True
        return False
    def __ge__(self,other):
        for index in range(self.lenth):
            if self.version_list[index] < other.version_list[index]:
                return False
        return True
class GamePlatformUpdater(object):
    def __init__(self,force=False):
        self.is_force = force
        self.zipbin      = '' # 7z的位置，pathlike
        self.workpath    = '' # 临时目录，pathlike
        self.btsync_path = '' # 目标目录，abs_pathlike
        self.ini_name    = '' # 配置文件所在目录，pathlike
        self.black_type  = '' # 类型黑名单，逗号分割。转成list使用
        self.CD = os.path.split(os.path.realpath(__file__))[0]
        self.config = configparser.ConfigParser()
        self.config.read(f'{self.CD}\\config.ini',encoding='utf-8')
        self.platform_list = self.config.sections()
        for key in self.config['updater']:
            self.__setattr__(key,self.config['updater'][key])
        self.platform_list.remove('updater')
class PlatformUtils(object):
    _instance_lock = threading.Lock()
    def __new__(cls ,*args, **kwargs):
        if not hasattr(cls,'_instance') :
            with PlatformUtils._instance_lock:
                if not hasattr(cls,'_instance'):
                    PlatformUtils._instance = super().__new__(cls)
        return PlatformUtils._instance

    def __init__(self,info:dict):
        self.src_info        = info
        self.check_file      = '' # 需要检测的文件，abs_pathlike
        self.src_folder      = '' # 需要打包的文件目录，逗号分割的str。转成list使用，abspath
        self.check_type      = '' # 检测版本信息，冒号分割。转成dict使用，key为检测的类型，value为检测的值。
        self.black_list      = '' # 目录黑名单，逗号分割。转成list使用，存在
        self.zip_folder_name = '' # 压缩文件夹名称
        self.zip_name        = '' # 压缩包名称
        self.addition_cmd    = '' # 需要执行的额外命令，逗号分割
        self.check_type      = '' # 检查类型
        self.check_ver       = '' # 版本号
        self.path_map = {
                            '$ALLUSERSPROFILE$'                   :   'C:\\ProgramData',
                            '$APPDATA$'                           :   'C:\\Users\\Admin\\AppData\\Roaming',
                            '$COMMONPROGRAMFILES$'                :   'C:\\Program Files\\Common Files',
                            '$COMMONPROGRAMFILES(x86)$'           :   'C:\\Program Files (x86)\\Common Files',
                            '$COMSPEC$'                           :   'C:\\Windows\\System32\\cmd.exe',
                            '$HOMEDRIVE$'                         :   'C:\\',
                            '$SystemDrive$'                       :   'C:\\',
                            '$HOMEPATH$'                          :   'C:\\Users\\Admin',
                            '$LOCALAPPDATA$'                      :   'C:\\Users\\Admin\\AppData\\Local',
                            '$PROGRAMDATA$'                       :   'C:\\ProgramData',
                            '$PROGRAMFILES$'                      :   'C:\\Program Files',
                            '$PROGRAMFILES(X86)$'                 :   'C:\\Program Files (x86)',
                            '$PUBLIC$'                            :   'C:\\UsersPublic',
                            '$SystemRoot$'                        :   'C:\\Windows',
                            '$TEMP$'                              :   'C:\\Users\\Admin\\AppData\\LocalTemp',
                            '$TMP$'                               :   'C:\\Users\\Admin\\AppData\\LocalTemp',
                            '$USERPROFILE$'                       :   'C:\\Admin',
                            '$WINDIR$'                            :   'C:\\Windows',
        }
        for key in info:
            if key in ['check_file']:
                value = info[key]
                if '$' in value :
                    replace_path_list = re.findall('\$.+\$',value)
                    for p in replace_path_list :
                        value = value.replace(p,self.path_map[p.upper()])
                self.__setattr__(key,value)
            elif key in ['src_folder','addition_cmd','black_list']:
                values = info[key].split(',')
                for value in values:
                    if '$' in value :
                        replace_path_list = re.findall('\$.+\$',value)
                        for p in replace_path_list :
                            value = value.replace(p,self.path_map[p.upper()])
                self.__setattr__(key,values)
            elif key in ['check_type']:
                value_list = info[key].split(':')
                self.__setattr__('check_type',value_list[0])
                self.__setattr__('check_ver',VersionType(value_list[1]))
            else:
                self.__setattr__(key,info[key])
    
    def need_update(self) -> bool:
        cf = self.check_file.replace('\\','\\\\')
        wmic_cmd = f'wmic datafile where name="{cf}" get version'
        ret_list = subprocess.Popen(wmic_cmd,creationflags=0x08000000,stdout=subprocess.PIPE).stdout.readlines()
        if len(ret_list) > 1 :
            local_version = VersionType(ret_list[1].strip().decode())
        if self.check_ver <= local_version:
            return False
        else:
            return True

    def run_platform(self) -> None:
        subprocess.call()
class DbgViewHandler(logging.Handler):
    def emit(self, record):
        OutputDebugString = ctypes.windll.kernel32.OutputDebugStringW
        OutputDebugString(self.format(record))
def get_logger() ->logging.Logger:
    logger = logging.getLogger(name='Platform_Updater')
    logger.setLevel(logging.DEBUG)
    logfmt = logging.Formatter("LASU : %(asctime)s\t%(levelname)s\t\t%(name)s\t\t\t%(message)s")
    # logging to dbgview
    ods = DbgViewHandler()
    ods.setLevel(logging.NOTSET)
    ods.setFormatter(logfmt)
    logger.addHandler(ods)
    # logging to console
    logstreamfd = logging.StreamHandler()
    logstreamfd.setLevel(logging.NOTSET)
    logstreamfd.setFormatter(logfmt)
    logger.addHandler(logstreamfd)
    return logger

if __name__ == "__main__":
    logger = get_logger()
    GPU = GamePlatformUpdater()
    for p_name in GPU.platform_list:
        PU = PlatformUtils(GPU.config[p_name])
        if not PU.need_update():
            logger.debug(f'dont need update this platform : {p_name}')
            continue
        logger.debug(f'will update this platform : {p_name}')
        