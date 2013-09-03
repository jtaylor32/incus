package main

import (
    "net/http"
    "log"
    "time"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    _ "net/http/pprof"
)

var DEBUG bool
var CLIENT_BROAD bool

func main() {
    defer func() {
        if err := recover(); err != nil {
            log.Println("FATAL: ", err)

            log.Println("clearing redis memory...")
            log.Println("Shutting down...")
        }
    }()

    conf  := initConfig()
    store := initStore(&conf)
    initLogger(conf)

    go func() {
        for {
            log.Println(store.memory.clientCount)
            time.Sleep(20 * time.Second)
        }
    }()
    
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)
    InstallSignalHandlers(signals)
    
    CLIENT_BROAD = conf.GetBool("client_broadcasts")
    server := createServer(&conf, &store)
    
    go server.initAppListener()
    go server.initSocketListener()
    go server.initPingListener()
    
    listenAddr := fmt.Sprintf(":%s", conf.Get("listening_port"))
    err := http.ListenAndServe(listenAddr, nil)
    if err != nil {
        log.Fatal(err)
    }
}

func InstallSignalHandlers(signals chan os.Signal) {
    go func() {
        sig := <-signals
        switch sig {
        case syscall.SIGKILL: fallthrough;
        case syscall.SIGINT:
            log.Println("\nCtrl-C signalled\n")
            os.Exit(0)
        }
    }()
}

func initLogger(conf Configuration) {
    fi, err := os.OpenFile("/var/log/incus.log", os.O_RDWR|os.O_APPEND, 0660);
    if err != nil {
        log.Fatalf("Error: %v", err.Error());
    }
    
    log.SetOutput(fi)
    
    DEBUG = false
    if conf.Get("log_level") == "debug" {
        DEBUG = true
    }
}