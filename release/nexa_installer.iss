#define MyAppName "Nexa-CLI"
#define MyAppVersion GetEnv('VERSION')
#define MyAppPublisher "Nexa AI"
#define MyAppExeName "nexa.exe"
#define MyAppServiceName "NexaService"
#define MyAppLauncherName "Nexa CLI Launcher.exe"

[Setup]
AppId={{e9b30237-d65d-4a79-a7c0-f4e217e78f54}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
OutputDir=..\artifacts
OutputBaseFilename=nexa-cli_windows-setup
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes
SetupIconFile=nexa_logo.ico
PrivilegesRequired=admin

[Files]
; Main executables
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cpu\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsCPUSelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cpu\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsCPUSelected

Source: "..\artifacts\nexa-cli_windows_llama-cpp-cuda\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsCUDASelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cuda\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsCUDASelected

Source: "..\artifacts\nexa-cli_windows_llama-cpp-vulkan\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsVulkanSelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-vulkan\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsVulkanSelected

Source: "..\artifacts\nexa-cli-launcher.exe"; DestDir: "{app}"; DestName: "{#MyAppLauncherName}"; Flags: ignoreversion
Source: "..\artifacts\nssm.exe"; DestDir: "{app}"; Flags: ignoreversion

; Dependencies - with corrected exclusions
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cpu\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsCPUSelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cuda\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsCUDASelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-vulkan\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsVulkanSelected

[Registry]
; Modified registry entries to properly handle icons and default applications
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath(ExpandConstant('{app}'))

; Launcher registration (primary application)
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppLauncherName}"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppLauncherName}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppLauncherName}"; ValueType: string; ValueName: "Path"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#MyAppName}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppLauncherName}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}\Shell\Open\Command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#MyAppLauncherName}"" ""%1"""; Flags: uninsdeletekey

; CLI executable registration (secondary)
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppExeName}"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppExeName}"; ValueType: string; ValueName: "Path"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppExeName}"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#MyAppName} CLI"; Flags: uninsdeletekey
Root: HKLM; Subkey: "SOFTWARE\Classes\Applications\{#MyAppExeName}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName}"; Flags: uninsdeletekey

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

[Run]
; Create and start the service using NSSM (Non-Sucking Service Manager)
; NSSM can properly wrap console applications as Windows services
Filename: "{app}\nssm.exe"; Parameters: "install ""{#MyAppServiceName}"" ""{app}\{#MyAppExeName}"" serve"; Flags: runhidden; Tasks: runasservice
Filename: "{app}\nssm.exe"; Parameters: "set ""{#MyAppServiceName}"" DisplayName ""Nexa AI SDK Service"""; Flags: runhidden; Tasks: runasservice
Filename: "{app}\nssm.exe"; Parameters: "set ""{#MyAppServiceName}"" Description ""Nexa AI SDK Background Service"""; Flags: runhidden; Tasks: runasservice
Filename: "{app}\nssm.exe"; Parameters: "set ""{#MyAppServiceName}"" AppDirectory ""{app}"""; Flags: runhidden; Tasks: runasservice
Filename: "{app}\nssm.exe"; Parameters: "set ""{#MyAppServiceName}"" Start SERVICE_AUTO_START"; Flags: runhidden; Tasks: runasservice
Filename: "{app}\nssm.exe"; Parameters: "start ""{#MyAppServiceName}"""; Flags: runhidden; Tasks: runasservice

[UninstallRun]
; Stop and delete the service upon uninstallation using NSSM
Filename: "{app}\nssm.exe"; Parameters: "stop ""{#MyAppServiceName}"""; Flags: runhidden
Filename: "{app}\nssm.exe"; Parameters: "remove ""{#MyAppServiceName}"" confirm"; Flags: runhidden
Filename: "taskkill.exe"; Parameters: "/F /IM nexa.exe"; Flags: runhidden
Filename: "taskkill.exe"; Parameters: "/F /IM nexa-cli.exe"; Flags: runhidden

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppLauncherName}"
Name: "{commondesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppLauncherName}"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
Name: "runasservice"; Description: "Run Nexa as a background service (runs 'nexa serve' on startup)"; GroupDescription: "Service Configuration"
