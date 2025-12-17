EXE     := .exe
RM      := powershell -NoProfile -NonInteractive "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -Path"
MKDIR   := powershell -NoProfile -NonInteractive "New-Item -ItemType Directory -Force -Path"
PYTHON  := python

.PHONY: build clean install-deps

build: install-deps
	@echo "====> Building executable with PyInstaller"
	$(MKDIR) dist
	$(PYTHON) -m PyInstaller \
		--name function-calling-demo \
		--onefile \
		--console \
		--clean \
		--distpath dist \
		--workpath build/pyinstaller \
		--hidden-import flask \
		--hidden-import mcp \
		--hidden-import nexaai \
		--hidden-import asyncio \
		--collect-all mcp \
		--collect-all nexaai \
		main.py
	@echo "====> Build complete: dist\function-calling-demo$(EXE)"

install-deps:
	@echo "====> Installing PyInstaller"
	$(PYTHON) -m pip install --upgrade pip
	$(PYTHON) -m pip install pyinstaller

clean:
	@echo "====> Cleaning build artifacts"
	-$(RM) build
	-$(RM) dist
	-$(RM) *.spec
