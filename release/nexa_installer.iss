#define MyAppName "Nexa-CLI"
#define MyAppVersion GetEnv('VERSION')
#define MyAppPublisher "Nexa AI"
#define MyAppExeName "nexa.exe"
#define MyAppServiceName "NexaService"
#define MyAppLauncherName "Nexa SDK Launcher.exe"
#define MyAppAliasCmdName "nexa-exe.cmd"

[Setup]
AppId={{e9b30237-d65d-4a79-a7c0-f4e217e78f54}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={localappdata}\{#MyAppName}
DefaultGroupName={#MyAppName}
OutputDir=artifacts
OutputBaseFilename=nexa-cli-windows-setup-${#MyAppVersion}
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes
SetupIconFile=nexa_logo.ico

[Files]
; Main executables
Source: "artifacts\nexasdk-cli_windows_llama-cpp-cpu\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsCPUSelected
Source: "artifacts\nexasdk-cli_windows_llama-cpp-cpu\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsCPUSelected

Source: "artifacts\nexasdk-cli_windows_llama-cpp-cuda\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsCUDASelected
Source: "artifacts\nexasdk-cli_windows_llama-cpp-cuda\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsCUDASelected

Source: "artifacts\nexasdk-cli_windows_llama-cpp-vulkan\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsVulkanSelected
Source: "artifacts\nexasdk-cli_windows_llama-cpp-vulkan\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsVulkanSelected

Source: "artifacts\nexa-cli-windows-launcher.exe"; DestDir: "{app}"; DestName: "{#MyAppLauncherName}"; Flags: ignoreversion

; Dependencies - with corrected exclusions
Source: "artifacts\nexasdk-cli_windows_llama-cpp-cpu\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsCPUSelected
Source: "artifacts\nexasdk-cli_windows_llama-cpp-cuda\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsCUDASelected
Source: "artifacts\nexasdk-cli_windows_llama-cpp-vulkan\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsVulkanSelected

[Registry]
; Modified registry entries to properly handle icons and default applications
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath(ExpandConstant('{app}'))

; Launcher registration (primary application)
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppLauncherName}"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppLauncherName}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppLauncherName}"; ValueType: string; ValueName: "Path"; ValueData: "{app}"
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#MyAppName}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppLauncherName}"
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}\Shell\Open\Command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#MyAppLauncherName}"" ""%1"""

; CLI executable registration (secondary)
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppExeName}"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppExeName}"; ValueType: string; ValueName: "Path"; ValueData: "{app}"
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppExeName}"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#MyAppName} CLI"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppExeName}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName}"

[Code]
var
  VersionPage: TInputOptionWizardPage;

procedure InitializeWizard;
begin
  VersionPage := CreateInputOptionPage(wpWelcome,
    'Choose Version', 'Which version of Nexa-cli do you want to install?',
    'Please select the version you want to install, then click Next.',
    True, False);

  VersionPage.Add('CUDA (12.4.1 or higher)');
  VersionPage.Add('Vulkan (1.3.261.1 or higher)');
  VersionPage.Add('CPU');

  VersionPage.SelectedValueIndex := 0;
end;

function IsCUDASelected: Boolean;
begin
  Result := VersionPage.SelectedValueIndex = 0;
end;

function IsVulkanSelected: Boolean;
begin
  Result := VersionPage.SelectedValueIndex = 1;
end;

function IsCPUSelected: Boolean;
begin
  Result := VersionPage.SelectedValueIndex = 2;
end;

function NeedsAddPath(Param: string): Boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE,
    'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
    'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Uppercase(Param) + ';', ';' + Uppercase(OrigPath) + ';') = 0;
end;

procedure CreateAliasCmdFile();
var
  CmdFilePath: string;
  CmdFileContents: TStringList;
begin
  CmdFilePath := ExpandConstant('{app}\{#MyAppAliasCmdName}');
  CmdFileContents := TStringList.Create;
  try
    CmdFileContents.Add('@echo off');
    CmdFileContents.Add('"%~dp0\{#MyAppExeName}" %*');
    CmdFileContents.SaveToFile(CmdFilePath);
  finally
    CmdFileContents.Free;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    CreateAliasCmdFile();
  end;
end;

[Run]
; Create and start the service if the user selected the task
Filename: "{sys}\sc.exe"; Parameters: "create ""{#MyAppServiceName}"" binPath= ""{app}\{#MyAppExeName} serve"" start= auto"; Flags: runhidden; Tasks: runasservice
Filename: "{sys}\sc.exe"; Parameters: "start ""{#MyAppServiceName}"""; Flags: runhidden; Tasks: runasservice

[UninstallRun]
; Stop and delete the service upon uninstallation
Filename: "{sys}\sc.exe"; Parameters: "stop ""{#MyAppServiceName}"""; Flags: runhidden
Filename: "{sys}\sc.exe"; Parameters: "delete ""{#MyAppServiceName}"""; Flags: runhidden

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppLauncherName}"
Name: "{commondesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppLauncherName}"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "runasservice"; Description: "Run Nexa as a background service (runs 'nexa serve' on startup)"; GroupDescription: "Service Configuration"; Flags: unchecked
