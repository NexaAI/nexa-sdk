import argparse
import sys
import shutil
import os
import subprocess

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
    parser = argparse.ArgumentParser(description="Build wheel for different platforms with GPU or CPU support.")
    parser.add_argument("--gpu", action="store_true", help="Use GPU for building")
    args = parser.parse_args()

    # Detect the operating system
    os_type = sys.platform

    # Determine if GPU or CPU version is being used
    version_type = "GPU" if args.gpu else "CPU"

    # Get the appropriate script and TOML file based on the OS and GPU flag
    toml_to_copy = toml_files[os_type][args.gpu]

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
        subprocess.run(["python", "-m", "build", "--wheel"], check=True)
        print(f"Build process completed successfully for {version_type} version on {os_type}.")
    except subprocess.CalledProcessError as e:
        print(f"Build failed: {e}")
        sys.exit(1)

    # Clean up the copied pyproject.toml file
    try:
        os.remove("pyproject.toml")
        print("Removed pyproject.toml")
    except FileNotFoundError:
        print("pyproject.toml file not found, skipping removal.")
    except OSError as e:
        print(f"Error removing pyproject.toml: {e}")

    # Optionally clean up the build directory
    try:
        shutil.rmtree("build")
        print("Removed build directory")
    except FileNotFoundError:
        print("Build directory not found, skipping removal.")
    except OSError as e:
        print(f"Error removing build directory: {e}")

    print("Build completed and cleaned up successfully")