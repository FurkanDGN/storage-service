package util

import (
   "log"
   "os"
)

func CreateDir(dirName string) {
   if _, err := os.Stat(dirName); os.IsNotExist(err) {
      err := os.MkdirAll(dirName, 0755)
      if err != nil {
         log.Fatal(err)
      }
   }
}