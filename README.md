# 文件搜索器

只是一个 Go 语言练手项目，会被用在 [Live Photo Tools](https://github.com/YuleBest/LivePhotoTools) 上。

> 使用 Go 编写
> 一个可以指定后缀名、日期范围、特别关键词的文件搜索器
> 并发处理提升搜索速度

## 使用方法

> `[]` 内为必须提供的参数

```shell
file-searcher -i [目录] -s <后缀名> -d <日期(天内)> -k <关键词>
```

或

```shell
file-searcher --input [目录] --subname <后缀名> --date <日期(天内)> --keyword <关键词>
```

### 选项

- `-h` `--help`：帮助

- `-i <directory>` `--input <directory>`：指定要搜索的目录（`<directory>`）

- `-s <string>` `--subname <string>`：指定要搜索的后缀名（`*.<string>`）

- `-d <positive integer>` `--date <positive integer>`：指定要搜索的时间范围（`<positive integer>` 天内）

- `-k <string>` `--keyword <string>`：指定特别关键词（`*<string>*`）

## 注意事项

- 指定特别关键词后并不是筛选包含特别关键词的文件，而是包含特别关键词的文件在列表中的文件名会被以蓝色显示