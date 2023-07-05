package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const workerCount = 50

type Dis struct {
	x int
	y int
}

var (
	disPool = make(chan Dis, 10)
)

func factory(disList []Dis, done chan bool) {
	disLen := len(disList)
	for i := 0; i < disLen; i++ {
		disPool <- Dis{
			x: disList[i].x,
			y: disList[i].y,
		}
	}
	close(disPool)
	done <- true

}

func main() {
	e := echo.New()
	e.POST("/get", func(c echo.Context) error {
		dis := c.FormValue("dis")
		finalDis := strings.Split(dis, "-")
		var distList []Dis
		for _, v := range finalDis {
			disValue := strings.Split(v, ":")
			distList = append(distList, Dis{x: castStringToInt(disValue[0]), y: castStringToInt(disValue[1])})
		}

		done := make(chan bool)
		go factory(distList, done)
		runWorker(workerCount, c.Request().Context())
		<-done

		return c.String(http.StatusOK, "Success")
	})

	e.Logger.Fatal(e.Start(":1212"))
}

func runWorker(num int, ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Add(num)
	for i := 0; i < num; i++ {
		go worker(ctx, wg, i)
	}
	wg.Wait()
}

func worker(ctx context.Context, wg *sync.WaitGroup, workerNumber int) {
	defer wg.Done()
	for {
		dis, open := <-disPool
		if !open {
			return
		}

		ctx = context.WithValue(ctx, workerNumber, Dis{
			x: dis.x,
			y: dis.y,
		})
		calculateDis(ctx, dis, workerNumber)
	}

}

func calculateDis(ctx context.Context, dis Dis, workerNumber int) {
	x, y := 0, 0
	currentCtx, err := typeConverter(ctx.Value(workerNumber))
	if err != nil {
		return
	}
	if workerNumber == ctx.Value(workerNumber) {
		x = currentCtx.x
		y = currentCtx.y
	}

	wg := new(sync.WaitGroup)
	defer wg.Done()
	wg.Add(4)

	go func() {
		if dis.x >= x {
			moveUp(x, dis.x, "x")
		}
		wg.Wait()
	}()

	go func() {
		if dis.y >= y {
			moveUp(y, dis.y, "y")
		}
		wg.Wait()
	}()

	go func() {
		if dis.x < x {
			moveDown(dis.x, x, "x")
		}
		wg.Wait()
	}()

	go func() {
		if dis.y >= y {
			moveDown(dis.y, y, "y")
		}
		wg.Wait()
	}()
}

func moveUp(from int, to int, state string) {
	for i := from; i < to; i++ {
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("current %s:%d \n", state, i)
	}
}

func moveDown(from int, to int, state string) {
	for i := from; i > to; i-- {
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("current %s:%d \n", state, i)
	}
}

func castStringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func typeConverter[R Dis](data any) (R, error) {
	var result R
	b, err := json.Marshal(&data)
	if err != nil {
		return R{}, err
	}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return R{}, err
	}
	return result, err
}
