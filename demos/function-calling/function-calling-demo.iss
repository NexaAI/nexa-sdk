[Setup]
AppName=Function Calling Demo
AppVersion=1.0
AppPublisher=NexaAI
DefaultDirName={autopf}\NexaAI\FunctionCallingDemo
DefaultGroupName=NexaAI Function Calling Demo
OutputDir=build\installer
OutputBaseFilename=function-calling-demo-setup-arm64
Compression=lzma
SolidCompression=yes
ArchitecturesAllowed=arm64
ArchitecturesInstallIn64BitMode=arm64
PrivilegesRequired=lowest
SetupIconFile=
UninstallDisplayIcon={app}\function-calling-demo.exe

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "build\dist\function-calling-demo.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Function Calling Demo"; Filename: "{app}\function-calling-demo.exe"; Parameters: "--serve"
Name: "{group}\{cm:UninstallProgram,Function Calling Demo}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\Function Calling Demo"; Filename: "{app}\function-calling-demo.exe"; Parameters: "--serve"; Tasks: desktopicon

[Run]
Filename: "{app}\function-calling-demo.exe"; Parameters: "--serve"; Description: "{cm:LaunchProgram,Function Calling Demo}"; Flags: nowait postinstall skipifsilent

[Code]
function InitializeSetup(): Boolean;
begin
  Result := True;
  // Check if Node.js is installed (optional check)
  // The application will show an error if Node.js is not found when running
end;

