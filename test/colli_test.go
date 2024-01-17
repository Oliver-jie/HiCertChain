package main

import(
        "fmt"
        "math/rand"
        "strconv"
        "time"
)

func main(){
        var table [14776336]int
        total_colli := 0
        for i:=0;i<1000000;i++{
                fmt.Printf("开始第%d次循环:",i)
                cert1:=RandInt(1000000000000)
                cert1String := strconv.Itoa(cert1)
                fmt.Println(cert1String)

                pos := cert1 % 14776336
                fmt.Printf("放置的位置为:%d",pos)
                if table[pos]==0 {
                        table[pos]=1
                }else{
                        total_colli++
                }
        }
        fmt.Println(total_colli)
}

func RandInt(len int)int  {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(len)

}
