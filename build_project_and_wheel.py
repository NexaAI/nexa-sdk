import argparse
import sys
import shutil
import os
import subprocess

build_scripts = {
    "linux": {
        True: "build_scripts/build_linux_gpu.sh",
        False: "build_scripts/build_linux_cpu.sh"
    },
    "darwin": {
        True: "build_scripts/build_metal.sh",
        False: "build_scripts/build_metal.sh",
    },
    "win32": {
        True: "build_scripts/build_win_cuda.sh",
        False: "build_scripts/build_win_cpu.sh"
    }
}

toml_files = {
    "linux": {
        True: "build_scripts/pyproject_linux_cuda.toml",
        False: "build_scripts/pyproject_linux_cpu.toml"
    },
    "darwin": {
        True: "build_scripts/pyproject_macos_metal.toml",
        False: "build_scripts/pyproject_macos_metal.toml",
    },
    "win32": {
        True: "build_scripts/pyproject_win_cuda.toml",
        False: "build_scripts/pyproject_win_cpu.toml"
    }
}

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--gpu", action="store_true", help="Use GPU for building")
    parser.add_argument("--os", type=str, choices=["linux", "darwin", "win32"], help="Specify the operating system")
    args = parser.parse_args()

    # Detect the operating system
    os_type = args.os if args.os else sys.platform

    if os_type not in build_scripts:
        print(f"Unsupported operating system: {os_type}")
        sys.exit(1)

    # Determine if GPU or CPU version is being used
    version_type = "GPU" if args.gpu else "CPU"

    # Get the appropriate script and TOML file based on the OS and GPU flag
    script_to_copy = build_scripts[os_type][args.gpu]
    toml_to_copy = toml_files[os_type][args.gpu]
    print(f"Script to copy: {script_to_copy}")
    print(f"TOML to copy: {toml_to_copy}")
    
    # Copy the selected script to the root directory
    try:
        shutil.copy(script_to_copy, "build_project.py")
        print(f"Copied {script_to_copy} ({version_type} version) to build_project.py for {os_type}")
    except IOError as e:
        print(f"Failed to copy {script_to_copy}: {e}")
        sys.exit(1)

    # Copy the selected TOML file to the root directory
    try:
        shutil.copy(toml_to_copy, "pyproject.toml")
        print(f"Copied {toml_to_copy} ({version_type} version) to pyproject.toml for {os_type}")
    except IOError as e:
        print(f"Failed to copy {toml_to_copy}: {e}")
        sys.exit(1)

    # Run the build process
    try:
        print(f"Starting the build process using {version_type} version on {os_type}...")
        subprocess.run(["python", "-m", "build"], check=True)
        print(f"Build process completed successfully for {version_type} version on {os_type}.")
    except subprocess.CalledProcessError as e:
        print(f"Build failed: {e}")
        sys.exit(1)

    # Clean up the copied files
    for file in ["pyproject.toml", "build_project.py"]:
        try:
            os.remove(file)
            print(f"Removed {file}")
        except FileNotFoundError:
            print(f"File {file} not found, skipping removal.")
        except OSError as e:
            print(f"Error removing {file}: {e}")

    # Optionally clean up the build directory
    try:
        shutil.rmtree("build")
        print("Removed build directory")
    except FileNotFoundError:
        print("Build directory not found, skipping removal.")
    except OSError as e:
        print(f"Error removing build directory: {e}")

    print("Build completed and cleaned up successfully")