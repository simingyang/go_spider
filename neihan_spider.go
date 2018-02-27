package main

import "fmt"
import "net/http"
import "log"
import "io/ioutil"
import "strconv"
import "regexp"
import "os"
import iconv "github.com/djimenez/iconv-go"

//明确目标
//第1页  https://www.neihanba.com/dz/index.html
//第2页  https://www.neihanba.com/dz/list_2.html
//第n页 https://www.neihanba.com/dz/list_n.html

//1 首先进入某页的页码主页，----> 取出每个段子链接地址
// https://www.neihanba.com 拼接一个段子的完整url路径
//得到每个段子路径的正则表达式  `<h4> <a href="(?:s(.*?))"`

// https://www.neihanba.com + /dz/1092886.html

// 进入每个段子的首页，得到段子的标题和内容

//标题的正则
//`<h1>(?:s(.*?))</h1>`

//内容的正则
//`<td><p>(?s:(.*?))</p></td>`

type Spider struct {
	Page int //当前爬虫已经爬取到了第几页
}

func (this *Spider) Store_one_page(titles []string, contents []string) error {

	filename := "myDuanzi.txt"

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("open file myduanzi.txt error ", err)
		return err
	}
	defer f.Close()

	for i := 0; i < len(titles); i++ {
		f.WriteString("\n====================\n")
		f.WriteString(titles[i])
		f.WriteString("\n--------------------\n")
		f.WriteString(contents[i])
	}

	return nil
}

//爬取一个段子
func (this *Spider) Spider_one_DZ(url string) (dz_title string, dz_content string) {

	content, rcode := this.HttpGet(url)
	if rcode != 200 {
		fmt.Println("url = ", url, " error rcode = ", rcode)
		return "", ""
	}

	//得到标题
	title_exp := regexp.MustCompile(`<h1>(.*?)</h1>`)
	titles := title_exp.FindAllStringSubmatch(content, -1)
	//fmt.Println("titles = ", titles)
	for _, title := range titles {
		dz_title = title[1]
		break
	}

	//得到内容
	content_exp := regexp.MustCompile(`<td><p>(?s:(.*?))</p></td>`)
	contents := content_exp.FindAllStringSubmatch(content, -1)
	//fmt.Println("contents= ", contents)
	for _, content := range contents {
		dz_content = content[1]
		break
	}

	return
}

//爬取一个某页的菜单页码
func (this *Spider) Spider_one_page() {
	fmt.Println("正在爬取 ", this.Page, " 页")

	url := ""

	if this.Page == 1 {
		url = "https://www.neihanba.com/dz/index.html"
	} else {
		url = "https://www.neihanba.com/dz/list_" + strconv.Itoa(this.Page) + ".html"
	}

	//fmt.Println(url)

	//爬取第一页
	content, rcode := this.HttpGet(url)
	if rcode != 200 {
		fmt.Println("http get error rcode = ", rcode)
		return
	}

	//得到每个段子的url路径
	dz_url_exp := regexp.MustCompile(`<h4> <a href="(.*?)"`)
	urls := dz_url_exp.FindAllStringSubmatch(content, -1)

	//fmt.Println(urls)

	//存储当前页面的全部标题和文本的slice
	title_slice := make([]string, 0)
	content_slice := make([]string, 0)

	for _, dz_url := range urls {
		full_url := "https://www.neihanba.com" + dz_url[1]

		fmt.Println(full_url)

		//根据段子的url路径 爬取每个段子的内容
		dz_title, dz_content := this.Spider_one_DZ(full_url)

		title_slice = append(title_slice, dz_title)
		content_slice = append(content_slice, dz_content)
	}

	//fmt.Println(title_slice)
	//将本页的段子存到文件里
	this.Store_one_page(title_slice, content_slice)

}

//请求一个页码将页码中的全部源码content
func (this *Spider) HttpGet(url string) (content string, statusCode int) {

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		content = ""
		statusCode = -100
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		statusCode = resp.StatusCode
		content = ""
		return
	}

	//将gb2312--->utf-8  data--->out
	out := make([]byte, len(data))
	out = out[:]

	iconv.Convert(data, out, "gb2312", "utf-8")

	content = string(out)
	statusCode = resp.StatusCode

	return
}

func (this *Spider) DoWork() {

	fmt.Println("Spider begin to  work")
	this.Page = 1

	var cmd string

	for {
		fmt.Println("请输入任意键爬取下一页，输入exit退出")
		fmt.Scanf("%s", &cmd)
		if cmd == "exit" {
			fmt.Println("exit")
			break
		}

		//需要爬取下一页
		this.Spider_one_page()

		this.Page++

	}

}

func main() {

	sp := new(Spider)
	sp.DoWork()

}
