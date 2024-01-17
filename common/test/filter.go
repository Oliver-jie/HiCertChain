	package main

	import(
		"fmt"
		"math/rand"
	//	"strconv"
		"time"
	)



	func main () {
		//设计布隆过滤器 一共有16个哈希计算方式，
		//设计基本的参数
		total_colli := 0

		//设计一个具有很高长度的二维数组 4096*16*16=1048576
		var table [32]int

		// start := time.Now()
		//开始循环的插入数据
		for i:=0;i<32;i++{
			table[i]=RandInt(16384)
		}
		for j := 0; j < 1000000; j++ {
			fmt.Printf("第%d次验证\n",j)
			tmp:=RandInt(16384)
			for k:=0;k<32;k++{
				if table[k]== tmp{
					total_colli++
					break
				}
			}
		}

		// elapsed := time.Since(start).Nanoseconds()
		

		//fmt.Printf("总共的假阳性次数为：%d\n",total_colli)
		fmt.Printf("碰撞的次数为%d\n",total_colli)
	}
	// //哈希计算
	// func compu_chart(cert1String string,num int)string{
	// 	cert2Hash :=MD5Hash2(cert1String)
	// 	// hash截取
	// 	cert3String := cert2Hash[2:10]
	// 	// 转换成十进制
	// 	chart:=Hex2Dec2(cert3String)%4096
	// 	// 指纹信息
	// 	cert2String := cert2Hash[0:8]
	// 	fp := (Hex2Dec2(cert2String))%16384
	// 	chart2 := fp^chart
	// 	if num == 1 {
	// 		return strconv.Itoa(chart)
	// 	}
	// 	if num == 2{
	// 		return strconv.Itoa(chart2%4096)
	// 	}
	// 	return strconv.Itoa(fp)+"--"+strconv.Itoa(chart)+"--"+strconv.Itoa(chart2)
	// }



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