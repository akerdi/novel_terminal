终端小说

使用cobra colly redis xorm 等构建

# Enter Novel

    novel [find/list/read]

Process is entered and listening for commands

## Usage

### Common Usage

#### find ${novelname} [x] 已实现

If local storage exist the novelname, screen prints the novelname:

《斗罗大陆》

Or this will list sites which novel do support. Highlight the first site if sites.

User use ↑ or ↓ to select target site, [novel] then will do the rest job for you.

#### delete ${novelname} [ ] 未实现

if local storage exist ${novelname}, delete it.

Storage will persist only one copy of a ${novelname}, so if you what to reselect site, user should delete ${novelname} first

#### list ${novelname} [x] 已实现

${novelname} optional.

Listing all you were in your local storage. such as:

《斗罗大陆》《神墓》《神雕侠侣》《倚天屠龙记》

User use ← or → to enter the novel, then screen listing all captors:

《第一章》《第二章》《第三章》

Select ← → ↑ ↓ to enter the captor.

In the captor, use ↑ ↓ to choose preview or next captor




# Usage 1.0

# 查看帮助

./novel (windows: .\novel.exe)

# 搜索书目

./novel find (windows: .\novel.exe find)

### 指定搜索书目

./novel find --novelname 斗罗大陆 (windows: .\novel.exe find --novelname 斗罗大陆)

# 列出本地书目

./novel list (windows: .\novel.exe list)

### 指定列出本地书目

./novel list --novelname 斗罗 (windows: .\novel.exe list --novelname 斗罗)

# 直接阅读本地书目

./novel read (windows: .\novel.exe read)

### 指定阅读本地书目

./novel read --novelname 斗罗 (windows: .\novel.exe read --novelname 斗罗)

---
---
---
---
> 阅读时: [上一页 a+Enter] [下一页 d+Enter] [返回选取章节 q+Enter] [结束程序 Ctrl+c]


[参考 YourNovel](https://github.com/DemonFengYuXiang/YourNovel)