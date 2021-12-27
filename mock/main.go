package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/cihub/seelog"
)

var nginxLog = `111.49.69.172 - [111.49.69.172] - - [18/Jun/2019:15:48:47 +0800] "GET /%s HTTP/1.1" 308 171 "-" "%s" %d 0.000 [default-http-svc3-80] - - - - %s %s
`

var javaStackLogCount = 3
var javaStackLog = `[%s] %d "BLOCKED_TEST pool-1-thread-1" prio=6 tid=0x0000000006904800 nid=0x28f4 runnable [0x000000000785f000]
java.lang.Thread.State: RUNNABLE
			 at java.io.FileOutputStream.writeBytes(Native Method)
			 at java.io.FileOutputStream.write(FileOutputStream.java:282)
			 at java.io.BufferedOutputStream.flushBuffer(BufferedOutputStream.java:65)
			 at java.io.BufferedOutputStream.flush(BufferedOutputStream.java:123)
			 - locked <0x0000000780a31778> (a java.io.BufferedOutputStream)
			 at java.io.PrintStream.write(PrintStream.java:432)
			 - locked <0x0000000780a04118> (a java.io.PrintStream)
			 at sun.nio.cs.StreamEncoder.writeBytes(StreamEncoder.java:202)
			 at sun.nio.cs.StreamEncoder.implFlushBuffer(StreamEncoder.java:272)
			 at sun.nio.cs.StreamEncoder.flushBuffer(StreamEncoder.java:85)
			 - locked <0x0000000780a040c0> (a java.io.OutputStreamWriter)
			 at java.io.OutputStreamWriter.flushBuffer(OutputStreamWriter.java:168)
			 at java.io.PrintStream.newLine(PrintStream.java:496)
			 - locked <0x0000000780a04118> (a java.io.PrintStream)
			 at java.io.PrintStream.println(PrintStream.java:687)
			 - locked <0x0000000780a04118> (a java.io.PrintStream)
			 at com.nbp.theplatform.threaddump.ThreadBlockedState.monitorLock(ThreadBlockedState.java:44)
			 - locked <0x0000000780a000b0> (a com.nbp.theplatform.threaddump.ThreadBlockedState)
			 at com.nbp.theplatform.threaddump.ThreadBlockedState$1.run(ThreadBlockedState.java:7)
			 at java.util.concurrent.ThreadPoolExecutor$Worker.runTask(ThreadPoolExecutor.java:886)
			 at java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:908)
			 at java.lang.Thread.run(Thread.java:662)

Locked ownable synchronizers:
			 - <0x0000000780a31758> (a java.util.concurrent.locks.ReentrantLock$NonfairSync)

[%s] %d "BLOCKED_TEST pool-1-thread-2" prio=6 tid=0x0000000007673800 nid=0x260c waiting for monitor entry [0x0000000008abf000]
java.lang.Thread.State: BLOCKED (on object monitor)
			 at com.nbp.theplatform.threaddump.ThreadBlockedState.monitorLock(ThreadBlockedState.java:43)
			 - waiting to lock <0x0000000780a000b0> (a com.nbp.theplatform.threaddump.ThreadBlockedState)
			 at com.nbp.theplatform.threaddump.ThreadBlockedState\$2.run(ThreadBlockedState.java:26)
			 at java.util.concurrent.ThreadPoolExecutor$Worker.runTask(ThreadPoolExecutor.java:886)
			 at java.util.concurrent.ThreadPoolExecutor\$Worker.run(ThreadPoolExecutor.java:908)
			 at java.lang.Thread.run(Thread.java:662)

Locked ownable synchronizers:
			 - <0x0000000780b0c6a0> (a java.util.concurrent.locks.ReentrantLock$NonfairSync)

[%s] %d "BLOCKED_TEST pool-1-thread-3" prio=6 tid=0x00000000074f5800 nid=0x1994 waiting for monitor entry [0x0000000008bbf000]
java.lang.Thread.State: BLOCKED (on object monitor)
			 at com.nbp.theplatform.threaddump.ThreadBlockedState.monitorLock(ThreadBlockedState.java:42)
			 - waiting to lock <0x0000000780a000b0> (a com.nbp.theplatform.threaddump.ThreadBlockedState)
			 at com.nbp.theplatform.threaddump.ThreadBlockedState\$3.run(ThreadBlockedState.java:34)
			 at java.util.concurrent.ThreadPoolExecutor$Worker.runTask(ThreadPoolExecutor.java:886
			 at java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:908)
			 at java.lang.Thread.run(Thread.java:662)

Locked ownable synchronizers:
			 - <0x0000000780b0e1b8> (a java.util.concurrent.locks.ReentrantLock$NonfairSync)
`

var defaultConfig = `
<seelog type="asynctimer" asyncinterval="5000" minlevel="info" >
 <outputs formatid="common">
	 <rollingfile type="size" filename="%s" maxsize="%d" maxrolls="%d"/>
 </outputs>
 <formats>
	 <format id="common" format="%%Msg" />
 </formats>
</seelog>
`

var logger = seelog.Disabled

var stdoutFlag = flag.Bool("stdout", true, "")
var stderrFlag = flag.Bool("stderr", false, "")
var filePath = flag.String("path", "", "")
var perLogFileSize = flag.Int("log-file-size", 20*1024*1024, "")
var maxLogFileCount = flag.Int("log-file-count", 10, "")
var logsPerSec = flag.Int("logs-per-sec", 1, "")
var logType = flag.String("log-type", "java", "nginx java random json")
var logErrType = flag.String("log-err-type", "random", "nginx java random json")
var totalCount = flag.Int("total-count", 100, "")
var itemLen = flag.Int("item-length", 100, "")
var keyCount = flag.Int("key-count", 10, "")

var nowCount = 0

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func mockJsonLog() string {
	kv := make(map[string]string)
	for i := 0; i < *keyCount; i++ {
		kv[RandString(10)] = RandString(*itemLen)
	}
	val, _ := json.Marshal(kv)
	val = append(val, '\n')
	return string(val)
}

func mockOneLog(timeStr, logType string) string {
	nowCount++
	switch logType {
	case "nginx":
		return fmt.Sprintf(nginxLog, RandString(16), RandString(16), nowCount, RandString(16), RandString(*itemLen))
	case "java":
		return fmt.Sprintf(javaStackLog, timeStr, nowCount, timeStr, nowCount, timeStr, nowCount)
	case "json":
		return mockJsonLog()
	}
	return fmt.Sprintf("%s %d %s\n", timeStr, nowCount, RandString(*itemLen))
}

func dumpOneLog(timeStr string) {
	if len(*filePath) > 0 {
		logger.Info(mockOneLog(timeStr, *logType))
	}
	if *stdoutFlag {
		os.Stdout.WriteString(mockOneLog(timeStr, *logType))
	}
	if *stderrFlag {
		os.Stderr.WriteString(mockOneLog(timeStr, *logErrType))
	}
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	if len(*filePath) > 0 {
		log.Println("use file output, path : ", *filePath)
		logConfig := fmt.Sprintf(defaultConfig, *filePath, *perLogFileSize, *maxLogFileCount)
		fmt.Println("log config, : " + logConfig)
		var err error
		logger, err = seelog.LoggerFromConfigAsString(logConfig)
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < *totalCount; {
		startTime := time.Now()
		timeStr := startTime.Format(time.RFC3339Nano)
		for j := 0; j < *logsPerSec; j++ {
			dumpOneLog(timeStr)
			i++
		}
		endTime := time.Now()
		time.Sleep(time.Second - endTime.Sub(startTime))
	}

}
