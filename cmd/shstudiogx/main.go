// Command shstudiogx is Shaiya Studio GX, the Shaiya data editor.
package main

import (
	"runtime"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/app"
)

func init() { runtime.LockOSThread() }

func main() { app.Run() }
