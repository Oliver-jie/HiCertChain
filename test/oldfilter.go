package main

import(
	"fmt"
	"math/rand"
//	"strconv"
	"time"
)



func main () {
	//设计布隆过滤器

	//设计基本的参数
	total_colli := 0

	//设计一个具有很高长度的二维数组
	var table [16000]int

	start := time.Now()
	//开始循环的插入数据
	for i:=0;i<20000;i++{
		//fmt.Printf("开始第%d次循环:",i)
		cert1:=RandInt(10000000000)
		//cert1String := strconv.Itoa(cert1) + RandString2(6)
		//fmt.Println(cert1String)

		pos := cert1 % 16000
		if table[pos]==0 {
			table[pos]=1
		}else{
			total_colli++
		}
	}
	elapsed := time.Since(start).Nanoseconds()
	

	//fmt.Printf("总共的假阳性次数为：%d\n",total_colli)
	fmt.Printf("代码运行的总时间为%d\n",elapsed)
}




func RandInt(len int)int  {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(len)
	
}

func RandString2(len int) string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        b := r.Intn(26) + 65
        bytes[i] = byte(b)
    }
    return string(bytes)
}