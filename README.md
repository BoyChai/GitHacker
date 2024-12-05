# GitHacker
## 帮助
```bash
PS E:\GoLang\GitHacker> .\GitHacker.exe -h
Usage of E:\GoLang\GitHacker\GitHacker.exe:
  -o string
        输出目录,默认值为当前位置的GitHacker_Output的主机地址目录下
  -t string
        指定类型,默认为url (default "url")
```
## 使用
```bash
.\GitHacker.exe http://127.0.0.1/.git/
```
如果本地有一个`.git`也可以通过`-t local`的方式来解析本地的