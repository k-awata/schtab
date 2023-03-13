@_default:
    just --list

# Release the binary to Github
release ver msg:
    dotnet publish
    cp bin/Release/net7.0/win-x64/publish/schtab.exe .
    rm -f *.zip
    7z a schtab_{{ver}}_win-x64.zip schtab.exe LICENSE README.md
    git tag -a v{{ver}} -m "{{msg}}"
    git push origin v{{ver}}
    gh release create -n "{{msg}}" v{{ver}} schtab_{{ver}}_win-x64.zip
