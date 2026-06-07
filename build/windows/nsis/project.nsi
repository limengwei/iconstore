Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows you to use this project.nsi manually
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
## !define INFO_COPYRIGHT      "(c) Now, My Company" # Default "© 2026, My Company"
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
!include "nsDialogs.nsh"
!include "LogicLib.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!define MUI_PAGE_CUSTOMFUNCTION_PRE dirPagePre
!define MUI_DIRECTORYPAGE_VARIABLE $INSTDIR
Var ADD_TO_PATH
Var INSTALL_DIR
Var CHECKBOX_PATH

Page custom pageOptions pageOptionsLeave
!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!define MUI_FINISHPAGE_RUN ""
!define MUI_FINISHPAGE_RUN_NOTCHECKED
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uninstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
   StrCpy $ADD_TO_PATH "1"
FunctionEnd

# --- Custom options page (install dir + add to PATH checkbox) ---
Function pageOptions
   !insertmacro MUI_HEADER_TEXT "Installation Options" "Choose installation directory and PATH settings"
   
   nsDialogs::Create 1018
   Pop $0
   
   ${NSD_CreateLabel} 0 0 100% 12u "Install to:"
   Pop $0
   
   ${NSD_CreateText} 0 14u 85% 14u $INSTDIR
   Pop $INSTALL_DIR
   
   ${NSD_CreateBrowseButton} 87% 14u 13% 14u "Browse..."
   Pop $0
   ${NSD_OnClick} $0 onBrowseDir
   
   ${NSD_CreateCheckBox} 0 40u 100% 12u "Add to System PATH (enables iconstore --mcp from terminal)"
   Pop $CHECKBOX_PATH
   ${NSD_SetState} $CHECKBOX_PATH ${BST_CHECKED}
   
   nsDialogs::Show
FunctionEnd

Function pageOptionsLeave
   ${NSD_GetText} $INSTALL_DIR $INSTDIR
   ${NSD_GetState} $CHECKBOX_PATH $1
   ${If} $1 == ${BST_CHECKED}
      StrCpy $ADD_TO_PATH "1"
   ${Else}
      StrCpy $ADD_TO_PATH "0"
   ${EndIf}
FunctionEnd

Function onBrowseDir
   nsDialogs::SelectFolderDialog "Select Installation Folder" $INSTDIR
   Pop $0
   ${If} $0 != "error"
      ${NSD_SetText} $INSTALL_DIR $0
   ${EndIf}
FunctionEnd

Function dirPagePre
   # Skip the default directory page since we have our custom one
   Abort
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    
    !insertmacro wails.files

    File "${ARG_ICONSTORE_DB}"

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    ; Add to PATH if checked
    ${If} $ADD_TO_PATH == "1"
       ReadRegStr $0 SHELL_CONTEXT "Environment" "Path"
       ; Check if already in PATH
       StrCpy $2 "$0"
       Push $2
       Push "$INSTDIR"
       Call StrContains
       Pop $3
       ${If} $3 == ""
          WriteRegExpandStr SHELL_CONTEXT "Environment" "Path" "$0;$INSTDIR"
          SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=1000
       ${EndIf}
    ${EndIf}
    
    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall" 
    !insertmacro wails.setShellContext

    ; Remove from PATH
    ReadRegStr $0 SHELL_CONTEXT "Environment" "Path"
    Push $0
    Push "$INSTDIR"
    Call un.RemoveFromPath
    Pop $0
    WriteRegExpandStr SHELL_CONTEXT "Environment" "Path" "$0"
    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=1000

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd

# --- String utility functions ---
Function StrContains
    Exch $R1 ; needle
    Exch
    Exch $R2 ; haystack
    Push $R3
    Push $R4
    StrLen $R3 $R1
    StrLen $R4 $R2
    ${While} $R4 >= $R3
        StrCpy $R0 $R2 $R3
        ${If} $R0 == $R1
            StrCpy $R0 "found"
            Goto done
        ${EndIf}
        StrCpy $R2 $R2 "" 1
        IntOp $R4 $R4 - 1
    ${EndWhile}
    StrCpy $R0 ""
    done:
    Pop $R4
    Pop $R3
    Pop $R2
    Pop $R1
    Exch $R0
FunctionEnd

Function un.RemoveFromPath
    Exch $R1 ; path to remove
    Exch
    Exch $R2 ; current PATH
    Push $R0
    Push $R3
    Push $R4
    Push $R5
    
    StrLen $R3 $R1
    StrLen $R4 $R2
    
    ${While} $R4 >= $R3
        StrCpy $R0 $R2 $R3
        ${If} $R0 == $R1
            ; Found - remove it
            StrCpy $R5 $R2 "" $R3
            StrCpy $R2 $R5
            ; Remove leading semicolon if present
            StrCpy $R5 $R2 1
            ${If} $R5 == ";"
                StrCpy $R2 $R2 "" 1
            ${EndIf}
            ; Remove trailing semicolon if present
            StrCpy $R5 $R2 "" -1
            ${If} $R5 == ";"
                StrCpy $R2 $R2 -1
            ${EndIf}
            Goto unDone
        ${EndIf}
        ; Advance by one character, looking for ;path
        StrCpy $R5 $R2 1
        ${If} $R5 == ";"
            StrCpy $R2 $R2 "" 1
            IntOp $R4 $R4 - 1
        ${Else}
            StrCpy $R2 $R2 "" 1
            IntOp $R4 $R4 - 1
        ${EndIf}
    ${EndWhile}
    
    unDone:
    Pop $R5
    Pop $R4
    Pop $R3
    Pop $R0
    Pop $R1
    Exch $R2
FunctionEnd
