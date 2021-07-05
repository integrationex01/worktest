package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Cook struct {
	totalcook int
	speedcook int
}
type Customer struct {
	totaleat int
	speedeat int
}

const (
	m = 3 //厨师数量
	n = 5 //顾客数量
)

var (
	N              = 30  //可存在的最大寿司量
	resource int64 = 100 //材料
	eok      int64 = 0   //顾客完成人数
	cok      int64 = 0   //厨师完成人数
	ok       int64 = 0   //通道开关指示器

)

var wg sync.WaitGroup

var cook = [m]Cook{
	{14, 7},
	{33, 1},
	{25, 1},
}
var customer = [n]Customer{
	{15, 18},
	{13, 3},
	{12, 16},
	{14, 17},
	{13, 10},
}

func main() {
	//创建 日志文档
	file, err := os.OpenFile("b.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	logger := log.New(file, "", log.LstdFlags|log.Llongfile)

	//创建有缓冲的通道传送寿司
	circle := make(chan int, N)

	//并发执行
	for i := 0; i < m; i++ {
		wg.Add(1)
		go makesushi(circle, cook, i, logger)
	}
	for j := 0; j < n; j++ {
		wg.Add(1)
		go eatsushi(circle, customer, j, logger)
	}

	wg.Wait()
	fmt.Println("今日营业结束")
	logger.Println("今日营业结束")

}

func makesushi(circle chan int, cook [m]Cook, i int, logger *log.Logger) {
	//只做传入操作

	//向main声明函数已结束
	defer wg.Done()

	for atomic.LoadInt64(&ok) == 0 {
		//判断条件，原料用完或者达成工作量
		if (atomic.LoadInt64(&resource) > 0) && (cook[i].totalcook > 0) {
			//模拟加工时间
			time.Sleep(time.Millisecond * time.Duration(cook[i].speedcook))

			//消耗资源
			atomic.AddInt64(&resource, -1)

			//向通道传入数据
			circle <- 1

			//工作量增加,工作总量减少
			cook[i].totalcook = cook[i].totalcook - 1

			//打印信息
			fmt.Printf("cook %d makes a new sushi,\t total is %d \n", i+1, len(circle))
			logger.Printf("cook %d makes a new sushi,\t total is %d \n", i+1, len(circle))
		} else {
			atomic.AddInt64(&cok, 1)
			fmt.Printf("厨师%d已做完工作量,下班.当前完成人数为: %d\n", i, atomic.LoadInt64(&cok))
			logger.Printf("厨师%d已做完工作量，下班\n", i)
			//若全部人都完成了工作量
			if atomic.LoadInt64(&cok) == n {
				atomic.AddInt64(&ok, 1)
			}
			break
		}
	}
	return

}

func eatsushi(circle chan int, customer [n]Customer, i int, logger *log.Logger) {
	//只做传出操作

	//向main声明函数已结束
	defer wg.Done()

	for atomic.LoadInt64(&ok) == 0 {
		//判断条件，达到饱食量
		if customer[i].totaleat > 0 {

			//模拟用餐时间
			time.Sleep(time.Millisecond * time.Duration(customer[i].speedeat))

			//通道传出数据，表示吃了一个
			<-circle

			//饱食度增加,总食量减少
			customer[i].totaleat = customer[i].totaleat - 2

			//打印信息
			fmt.Printf("customer %d eats a sushi,\t total is %d \n", i+1, len(circle))
			logger.Printf("customer %d eats a sushi,\t total is %d \n", i+1, len(circle))
		} else {
			atomic.AddInt64(&eok, 1)
			fmt.Printf("顾客%d已吃饱,离开.当前吃饱人数为: %d\n", i, eok)
			logger.Printf("顾客%d已吃饱\n", i)
			//全部人都吃饱了
			if atomic.LoadInt64(&eok) == n {
				fmt.Println("全员均已吃饱")
				atomic.AddInt64(&ok, 1)
			}
			break
		}
	}
	return

}
