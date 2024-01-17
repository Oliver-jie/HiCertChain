package main

import "fmt"
import "math"


func main () {
a1 :=0.1
a2 :=0.2
a3 :=0.2
a4 :=0.2
a5 :=0.2

k1:=0.6
k2:=0.1
k3:=0.1
k4:=0.1
k5:=0.1
//获取值
var arr1 [12]float64
var arr2 [12]float64
var arr3 [12]float64
var arr4 [12]float64
var arr5 [12]float64

var ayy1 [12]float64
var ayy2 [12]float64
var ayy3 [12]float64
var ayy4 [12]float64
var ayy5 [12]float64

var y [12]float64

n:=1

//矩阵X赋值
for i := 1.0; i < 11.0; i=i+1.0 {
	arr1[n]=math.Pow(a1,i)
	arr2[n]=a2*i
	arr3[n]=(a3-a3*math.Pow(a1,i))/(1-a1)
	arr4[n]=a4*i
	arr5[n]=a5*i
	n=n+1
}

max1:=maxnum(arr1)
min1:=minnum(arr1)
max2:=maxnum(arr2)
min2:=minnum(arr2)
max3:=maxnum(arr3)
min3:=minnum(arr3)
max4:=maxnum(arr4)
min4:=minnum(arr4)
max5:=maxnum(arr5)
min5:=minnum(arr5)

//矩阵Y
n=1
for j := 1; j < 11; j++ {
	ayy1[j]=(arr1[j]-min1)/(max1-min1)
	ayy2[j]=(arr2[j]-min2)/(max2-min2)
	ayy3[j]=(arr3[j]-min3)/(max3-min3)
	ayy4[j]=(arr4[j]-min4)/(max4-min4)
	ayy5[j]=(arr5[j]-min5)/(max5-min5)
}

//最终的对比
for k := 1; k < 11; k++ {
	y[k]=(math.Pow(ayy1[k],k1))*(math.Pow(ayy2[k],k2))*(math.Pow(ayy3[k],k3))*(math.Pow(ayy4[k],k4))*(math.Pow(ayy5[k],k5))
}

//输出
for m := 1; m < 11; m++ {
	fmt.Println(y[m])
}
}


func maxnum(arr [12]float64) (max float64) {
	max= arr[0]//假设数组的第一位为最大值
    //常规循环，找出最大值
	for i := 1; i < 11; i ++ {
		if max < arr[i]{
			max = arr[i]
		}
	}
	return max
}
 
func minnum(arr [12]float64) (min float64) {
	min= arr[0]//假设数组的第一位为最小值
    //for-range循环方式，找出最小值
	for i := 1; i < 11; i ++ {
		if min > arr[i]{
			min = arr[i]
		}
	}
	return min
}

