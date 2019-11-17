# lipdf

主要利用pdftk进行pdf的处理与填充

[pdftk文档及安装](https://www.pdflabs.com/tools/pdftk-the-pdf-toolkit/)

## 处理流程

1. 上传PDF文件
2. 利用pdftk.generate_fdf指令获取fdf文件，得到有效的表单field
3. 利用pdftk.dump_data_fields_utf8指令获取表单field对应的属性，包括类型（文本|按钮..）, 默认值等等
4. 获取表单详细信息(FieldName, FieldType)后转为json数据
5. 将json数据返回给前端，进行展示
6. 网页填写对应数据后，提交数据
7. 服务端获取数据后，转换成fdf文件，利用pdftk.fill_form写入对应的pdf
8. 下载新pdf文件即可得到填充后的文件


## 说明

### generate_fdf获取fdf文件并读取

pdf中可交互表单数据格式为FDF(Forms Data Format)，[FDF](https://en.wikipedia.org/wiki/PDF#Forms_Data_Format_.28FDF.29),

利用pdftk指令generate_fdf可以从PDF文件中导出fdf数据
```shell
pdftk in.pdf generate_fdf output out.fdf
```

导出后的fdf示例数据如下, 其中Fields表示填充字段，Kids表示子填充字段

```

%FDF-1.2
%âãÏÓ
1 0 obj 
<<
/FDF 
<<
/Fields [
<<
/Kids [
<<
/V /
/T (marital wid)
>> 
<<
/V ()
/T (ident cntry)
>>]
/T (ap)
>>]
>>
>>
endobj 
trailer

<<
/Root 1 0 R
>>
%%EOF
```

观察fdf数据，可以提取的有用信息如下, 每一层括号表示一个子类

```

[[marital wid, ident cntry]ap]

```

因此在 ```core.fdf.go```中```readFormFields```方法利用栈的对称性可以
获得fdf的有效key集合

```go

package core

import (
	"bufio"
	"container/list"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// read and parse pdf form field keys
func readFormFields(filePath string) (map[string]struct{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("fail to open file:%v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	l := list.New()
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		if strings.Contains(line, "[")  {
			l.PushBack("[")
		}
		if strings.Contains(line, "]")  {
			l.PushBack("]")
		}

		if strings.HasPrefix(line, "/T")  {
			l.PushBack(strings.TrimSuffix(strings.TrimPrefix(line, "/T ("), ")\n"))
		}
	}
	if err != io.EOF {
		return nil, fmt.Errorf("fail to read file:%v", err)
	}

	if l.Len() < 2 {
		return nil, nil
	}

	//trim first "["
	l.Remove(l.Front())
	//trim last "]"
	l.Remove(l.Back())

	keys := make(map[string]struct{})
	prefixes := make([]string, 0, 1)
	for l.Len() > 0 {
		str := l.Back().Value.(string)
		if str != "]" && str != "[" {

			//prev value
			prev := l.Back().Prev()
			if prev == nil {
				keys[str] = struct{}{}
				break
			}

			//last prefix
			prefix := ""
			if len(prefixes) > 0 {
				prefix = prefixes[len(prefixes)-1]
			}

			prevStr:= prev.Value.(string)
			if prevStr == "]" {
				//just prefix, don't need to add to keys
				if len(prefix) > 0 {
					prefix = fmt.Sprintf("%s.%s", prefix, str)
				}else {
					prefix = str
				}
				prefixes = append(prefixes, prefix)
			}else {

				// add to keys
				k := ""
				if len(prefix) > 0 {
					k = fmt.Sprintf("%s.%s", prefix, str)
				}else {
					k = str
				}
				keys[k] = struct{}{}

				// [[[a]b]c]
				if prevStr == "[" {
					for prevStr == "[" {
						prefixes = prefixes[0:len(prefixes)-1]
						prev = prev.Prev()
						if prev != nil {
							prevStr = prev.Value.(string)
						}else {
							prevStr = ""
						}
					}
				}
			}
		}
		l.Remove(l.Back())
	}

	return keys, nil
}

```

### dump_data_fields_utf8获取dump数据

利用pdftk指令dump_data_fields_utf8可以从PDF文件中导出dump_field数据
```shell
pdftk in.pdf dump_data_fields_utf8 output out.fdf
```

导出的数据示例如下：

```

---
FieldType: Text
FieldName: ap.info dtl 1
FieldFlags: 8392704
FieldJustification: Left
---
FieldType: Text
FieldName: ap.new dtl 1
FieldFlags: 8392704
FieldJustification: Left

```

### 利用pdftk.fill_form填充PDF表单

利用pdftk.fill_form指令，利用生成的fdf文件充PDF表单
```shell
pdftk form.pdf fill_form data.fdf output form.filled.pdf
```