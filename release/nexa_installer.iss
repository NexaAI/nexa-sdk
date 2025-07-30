#define MyAppName "Nexa CLI"
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
DefaultDirName={localappdata}\{#MyAppName}
DefaultGroupName={#MyAppName}
OutputDir=..\artifacts
OutputBaseFilename=nexa-cli_windows-setup
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes
SetupIconFile=nexa_logo.ico
PrivilegesRequired=lowest
UninstallDisplayName={#MyAppName}
UninstallDisplayIcon={app}\{#MyAppLauncherName}
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

[Files]
; Main executables
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cpu\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsCPUSelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cpu\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsCPUSelected

Source: "..\artifacts\nexa-cli_windows_llama-cpp-cuda\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsCUDASelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cuda\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsCUDASelected

Source: "..\artifacts\nexa-cli_windows_llama-cpp-vulkan\nexa.exe"; DestDir: "{app}"; DestName: "nexa.exe"; Flags: ignoreversion; Check: IsVulkanSelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-vulkan\nexa-cli.exe"; DestDir: "{app}"; DestName: "nexa-cli.exe"; Flags: ignoreversion; Check: IsVulkanSelected

Source: "..\artifacts\nexa-cli-launcher.exe"; DestDir: "{app}"; DestName: "{#MyAppLauncherName}"; Flags: ignoreversion

; Dependencies - with corrected exclusions
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cpu\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsCPUSelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-cuda\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsCUDASelected
Source: "..\artifacts\nexa-cli_windows_llama-cpp-vulkan\lib\*"; DestDir: "{app}\lib"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsVulkanSelected

[Registry]
; Launcher registration (primary application)
Root: HKCU; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppLauncherName}"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppLauncherName}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppLauncherName}"; ValueType: string; ValueName: "Path"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#MyAppName}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppLauncherName}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\Applications\{#MyAppLauncherName}\Shell\Open\Command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#MyAppLauncherName}"" ""%1"""; Flags: uninsdeletekey

; CLI executable registration (secondary)
Root: HKCU; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppExeName}"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\{#MyAppExeName}"; ValueType: string; ValueName: "Path"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\Applications\{#MyAppExeName}"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#MyAppName} CLI"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\Applications\{#MyAppExeName}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName}"; Flags: uninsdeletekey

[Code]
const
  EnvironmentKey = 'Environment';

var
  VersionPage: TInputOptionWizardPage;

function InitializeSetup(): Boolean;
var
  UninstallString: String;
  ResultCode: Integer;
begin
  Result := True;

  UninstallString := '';
  RegQueryStringValue(HKCU, ExpandConstant('Software\Microsoft\Windows\CurrentVersion\Uninstall\{#emit SetupSetting("AppId")}_is1'), 'UninstallString', UninstallString);

  if UninstallString <> '' then
  begin
    if MsgBox('Existing version detected.'#13#10 +
               'Please uninstall the existing version first.'#13#10#13#10 +
               'Uninstall now?', mbConfirmation, MB_YESNO) = IDYES then
    begin
      if not Exec(RemoveQuotes(UninstallString), '/SILENT', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
      begin
        MsgBox('Uninstall failed. Please try again later.', mbError, MB_OK);
        Result := False;
      end
      else if ResultCode <> 0 then
      begin
        MsgBox(Format('Uninstall failed (ErrCode: %d).', [ResultCode]), mbError, MB_OK);
        Result := False;
      end
      else
      begin
        MsgBox('Uninstall successful', mbInformation, MB_OK);
      end;
    end
    else
    begin
      MsgBox('Installation aborted.', mbInformation, MB_OK);
      Result := False;
    end;
  end;
end;

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

procedure EnvAddPath(Path: string);
var
    Paths: string;
begin
    if not RegQueryStringValue(HKCU, EnvironmentKey, 'Path', Paths)
    then Paths := '';

    if Pos(';' + Uppercase(Path) + ';', ';' + Uppercase(Paths) + ';') > 0 then exit;

    Paths := Paths + ';'+ Path +';'

    if RegWriteStringValue(HKCU, EnvironmentKey, 'Path', Paths)
    then Log(Format('The [%s] added to PATH: [%s]', [Path, Paths]))
    else Log(Format('Error while adding the [%s] to PATH: [%s]', [Path, Paths]));
end;

procedure EnvRemovePath(Path: string);
var
    Paths: string;
    P: Integer;
begin
    if not RegQueryStringValue(HKCU, EnvironmentKey, 'Path', Paths) then
        exit;

    P := Pos(';' + Uppercase(Path) + ';', ';' + Uppercase(Paths) + ';');
    if P = 0 then exit;

    Delete(Paths, P - 1, Length(Path) + 1);

    if RegWriteStringValue(HKCU, EnvironmentKey, 'Path', Paths)
    then Log(Format('The [%s] removed from PATH: [%s]', [Path, Paths]))
    else Log(Format('Error while removing the [%s] from PATH: [%s]', [Path, Paths]));
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
    if CurStep = ssPostInstall
     then EnvAddPath(ExpandConstant('{app}'));
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
    if CurUninstallStep = usPostUninstall
    then EnvRemovePath(ExpandConstant('{app}'));
end;

[UninstallRun]
Filename: "taskkill.exe"; Parameters: "/F /IM nexa.exe"; Flags: runhidden
Filename: "taskkill.exe"; Parameters: "/F /IM nexa-cli.exe"; Flags: runhidden

[UninstallDelete]
Type: files; Name: "{app}\*.exe"
Type: files; Name: "{app}\*.dll"
Type: filesandordirs; Name: "{app}\lib"
Type: dirifempty; Name: "{app}"

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppLauncherName}"
Name: "{userdesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppLauncherName}"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
