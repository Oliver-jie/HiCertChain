package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"io"
	"crypto/md5"
	"bytes"
	"encoding/binary"

)



func main () {
	//随机生成一个证书编号，一共有6位
    cert1 := rand.Intn(10000)
	cert1String := strconv.Itoa(cert1)
	cert2string :=MD5Hash(cert1String)
	cert1Byte := Int64ToBytes(int64(cert1))

	var i int =5
    fmt.Printf("%b\n",i)   //--->>显示5的二进制数
	fmt.Printf("%b\n",cert1)   //--->>显示5的二进制数

	te1:=2^3
	te2:=3^5
	fmt.Printf("%d\n",te1)
	fmt.Printf("%d\n",te2)


	//设计两种取哈希的办法，并且计算每一个证书编号的哈希值
	//开始设计两张表分别为表一和表二，设计表的大小规则
	//开始装填证书信息
	fmt.Printf("%d\n",cert1)
	fmt.Printf(cert1String)
	fmt.Printf(cert2string)
	fmt.Printf("%d\n",cert1Byte)
}


func MD5Hash(data string) string {
    t := md5.New()
    io.WriteString(t, data)
    return fmt.Sprintf("%x", t.Sum(nil))
}

func Int64ToBytes(num int64) []uint8 {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		fmt.Println("int64转[]uint8失败")
	}
	return buffer.Bytes()
   }