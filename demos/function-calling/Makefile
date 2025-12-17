EXE     := .exe
RM      := powershell -NoProfile -NonInteractive "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -Path"
MKDIR   := powershell -NoProfile -NonInteractive "New-Item -ItemType Directory -Force -Path"
PYTHON  := python

.PHONY: build clean install-deps package

build: install-deps
	@echo "====> Building executable with PyInstaller"
	$(MKDIR) build/dist
	$(PYTHON) -m PyInstaller \
		--name function-calling-demo \
		--onefile \
		--console \
		--clean \
		--distpath build/dist \
		--workpath build/pyinstaller \
		--hidden-import flask \
		--hidden-import mcp \
		--hidden-import nexaai \
		--hidden-import asyncio \
		--collect-all mcp \
		--collect-all nexaai \
		main.py
	@echo "====> Build complete: build\dist\function-calling-demo$(EXE)"

install-deps:
	@echo "====> Installing PyInstaller"
	$(PYTHON) -m pip install --upgrade pip
	$(PYTHON) -m pip install pyinstaller

package: build
	@echo "====> Packaging installer with InnoSetup (Windows ARM64)"
	$(MKDIR) build/installer
	@powershell -NoProfile -NonInteractive -File package.ps1
	@echo "====> Package complete: build\installer\function-calling-demo-setup-arm64.exe"

clean:
	@echo "====> Cleaning build artifacts"
	-$(RM) build
	-$(RM) *.spec
