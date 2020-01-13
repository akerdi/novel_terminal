终端小说

使用cobra colly redis xorm 等构建

# Enter Novel

    novel

Process is entered and listening for commands

## Usage

### Common Usage

    exit

process stop

    type (ESC)

Process command level will pop out one level

#### find ${novelname}

If local storage exist the novelname, screen prints the novelname:

《斗罗大陆》

Or this will list sites which novel do support. Highlight the first site if sites.

User use ↑ or ↓ to select target site, [novel] then will do the rest job for you.

#### delete ${novelname}

if local storage exist ${novelname}, delete it.

Storage will persist only one copy of a ${novelname}, so if you what to reselect site, user should delete ${novelname} first

#### list ${novelname}

${novelname} optional.

Listing all you were in your local storage. such as:

《斗罗大陆》《神墓》《神雕侠侣》《倚天屠龙记》

User use ← or → to enter the novel, then screen listing all captors:

《第一章》《第二章》《第三章》

Select ← → ↑ ↓ to enter the captor.

In the captor, use ↑ ↓ to choose preview or next captor




## Usage

novel find --novelname 一片雨

novel find
