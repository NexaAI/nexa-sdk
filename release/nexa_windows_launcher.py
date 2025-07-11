import subprocess

# Launch PowerShell with the -NoProfile flag which is generally faster to start
# Use -NoExit to keep PowerShell open after running the command
# Use -Command to specify the command to run
subprocess.Popen(
    ["powershell", "-NoProfile", "-NoExit", "-Command", "nexa"],
    creationflags=subprocess.CREATE_NEW_CONSOLE
)