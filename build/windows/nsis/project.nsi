# Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows you to use this file manually
## from outside of Wails for debugging and development of the installer.
## 
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the wails_tools.nsh file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "my-project" # Default "iconstore"
## !define INFO_COMPANYNAME    "My Company" # Default "My Company"
## !define INFO_PRODUCTNAME    "My Product Name" # Default "My Product"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "0.1.0"
## !define INFO_COPYRIGHT      "(c) Now, My Company" # Default "(c) 2026, My Company"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
## !define WAILS_INSTALL_SCOPE     "user"             # Default "machine" - set to "user" for per-user install ($LOCALAPPDATA) without UAC prompt
####
## Include the wails tools
####
!define INFO_COMPANYNAME    "MindSpace"
!define INFO_PRODUCTNAME    "IconStore"
!define INFO_PRODUCTVERSION "1.0.0"
!define INFO_COPYRIGHT      "(c) 2026, MindSpace"
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
!if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif
ShowInstDetails show

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

# --- Main section (always installed) ---
Section "!IconStore (required)" SEC_MAIN
    SectionIn RO

    !insertmacro wails.setShellContext
    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files
    File "${ARG_ICONSTORE_DB}"

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

# --- Optional: Add to PATH ---
Section "Add to System PATH" SEC_PATH
    ReadRegStr $0 SHELL_CONTEXT "Environment" "Path"
    ; Check if already in PATH
    StrLen $1 $INSTDIR
    StrCpy $2 0
    ${Do}
        StrCpy $3 $0 $1 $2
        ${If} $3 == $INSTDIR
            ${ExitDo}
        ${EndIf}
        IntOp $2 $2 + 1
        StrLen $4 $0
    ${LoopUntil} $2 >= $4

    ${If} $3 != $INSTDIR
        WriteRegExpandStr SHELL_CONTEXT "Environment" "Path" "$0;$INSTDIR"
        SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=1000
    ${EndIf}
SectionEnd

# --- Section descriptions ---
LangString DESC_SEC_MAIN ${LANG_ENGLISH} "IconStore application files (required)"
LangString DESC_SEC_PATH ${LANG_ENGLISH} "Add IconStore to system PATH so you can run 'iconstore --mcp' from terminal"

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
    !insertmacro MUI_DESCRIPTION_TEXT ${SEC_MAIN} $(DESC_SEC_MAIN)
    !insertmacro MUI_DESCRIPTION_TEXT ${SEC_PATH} $(DESC_SEC_PATH)
!insertmacro MUI_FUNCTION_DESCRIPTION_END

Section "uninstall"
    !insertmacro wails.setShellContext

    ; Remove from PATH
    ReadRegStr $R0 SHELL_CONTEXT "Environment" "Path"
    StrLen $R1 $INSTDIR
    StrCpy $R2 0
    ${Do}
        StrCpy $R3 $R0 $R1 $R2
        ${If} $R3 == $INSTDIR
            ; Found - remove it
            StrCpy $R4 $R0 $R2 0
            IntOp $R2 $R2 + $R1
            StrCpy $R5 $R0 "" $R2
            ; Remove leading semicolon from remainder
            StrCpy $R6 $R5 1
            ${If} $R6 == ";"
                StrCpy $R5 $R5 "" 1
            ${EndIf}
            ; Remove trailing semicolon from prefix
            StrLen $R6 $R4
            IntOp $R6 $R6 - 1
            ${If} $R6 > 0
                StrCpy $R6 $R4 1 $R6
                ${If} $R6 == ";"
                    StrCpy $R4 $R4 -1
                ${EndIf}
            ${EndIf}
            StrCpy $R0 "$R4$R5"
            ${ExitDo}
        ${EndIf}
        IntOp $R2 $R2 + 1
        StrLen $R6 $R0
    ${LoopUntil} $R2 >= $R6

    WriteRegExpandStr SHELL_CONTEXT "Environment" "Path" "$R0"
    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=1000

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"
    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
