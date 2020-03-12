#dev

go run main.go [find|list|read] [--novelname 霸道人生]

# build

## mac

        CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
        
 This create a file: novel, copy [novel & rule.json & dictionary.txt] in the same directory do the trick
 
## windows

Because novel use `go-sqlite3`, so Mac can't build windows by throw `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub`

Mac 先安装`mingw-w64`

        brew install mingw-64

        CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -v
        
## linux

        brew install FiloSottile/musl-cross/musl-cross

        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc CGO_LDFLAGS="-static" go build -a -v
        
-a: 重新编译

-static 标识静态链接，没有这个选项，linux上运行报: -bash: ./xxx: /lib/ld-musl-x86_64.so.1:bad ELF interpreter: No such file or directory

## Usage

### 注意事项

期间必须附带 **rule.json** 文件，和 **dictionary.txt** 文件，否则会报错

## mac

        ./novel [find|list|read] [--novelname yourNovelName]

## windows

        ./novel.exe [find|list|read] [--novelname yourNovelName]

## linux

        ./novel-linux [find|list|read] [--novelname yourNovelName]