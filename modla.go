package main

import (
	"bufio"
	"flag"
	"fmt"

	// "image/color"
	"log"
	"os"
	"os/exec"

	"github.com/chzyer/readline"
	"github.com/kballard/go-shellquote"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/pipe.v2"

	// "github.com/nsf/termbox-go"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

func main() {
	menusrcPtr := flag.String("menusrc", "", "executable file used as menu source")
	//recentdbPtr := flag.String("recentdb", "", "db file used for recent items")
	//initqueryPtr := flag.String("initquery", "", "initial query")
	previewLinePtr := flag.String("preview", "", "preview selected line from source")
	dbgPtr := flag.Bool("debug", false, "show executed command and params")

	flag.Parse()

	initquery := strings.Join(flag.Args(), " ")

	// if preview only formated print and exit
	if *previewLinePtr != "" {
		previewMenu(previewLinePtr)
		os.Exit(0)
	}

	// normal menu run
	if *menusrcPtr == "" {
		panic("Not defined menu source")
	}

	hostname, err2 := os.Hostname()
	if err2 != nil {
		fmt.Printf("%v\n", err2)
	}

	p := pipe.Line(
		pipe.Exec("cmd", "/C", *menusrcPtr),
		pipe.Exec("fzf",
			"--delimiter=\t", "--with-nth=1",
			"--multi", "--exact", "--query", initquery,
			"--prompt", hostname+", "+time.Now().Format("Mon 2.1. 15:04")+": ",
			"--reverse",
			"--border",
			// TODO Lebeda - vybrat automaticky
			"--preview", "fzfMenu -preview {}", // TODO doladit preview program
			"--preview-window", "down,10",
			"--expect=f1,f2,f3,f4,f5,f6,f7,f8,f9",
			"--no-sort",
			// "--tabstop=" + strconv.Itoa(w - 8),  TODO smazat
			"--ansi"),
	)

	output, err := pipe.CombinedOutput(p)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	outstr := string(output)
	lines := strings.Split(outstr, "\n")

	// read key
	key := lines[0]
	if *dbgPtr {
		fmt.Println("key: " + key)
	}

	for _, s2 := range lines[1:] {
		// execute
		cmds := strings.Split(s2, "\t")
		if len(cmds) > 1 { // cmd[0] is only description, not command, cmd[1] on enter, cmd[2-9] on f2-9
			//fmt.Println(cmds[1])
			cmdstr := strings.TrimSpace(cmds[1])

			// What key fzf exited
			if strings.HasPrefix(key, "f") {
				fKeyNumStr := strings.Replace(key, "f", "", 1)
				fKeyNum, err := strconv.Atoi(fKeyNumStr)
				if err != nil {
					panic(err)
				}

				if len(cmds) < fKeyNum {
					fKeyNum = len(cmds)
					fmt.Println("exec by " + key + " but cmd not found. Use max cmd " + strconv.Itoa(fKeyNum))
				}
				cmdstr = strings.TrimSpace(cmds[fKeyNum])
			}

			// replace paterns {input}
			if strings.Contains(cmdstr, "{input}") {
				fmt.Println("Input for command: ", cmdstr)
				rl, err := readline.New("> ")
				if err != nil {
					panic(err)
				}
				defer rl.Close()

				line, err := rl.Readline()
				strings.TrimSpace(line)

				cmdstr = strings.ReplaceAll(cmdstr, "{input}", line)
				// TODO Lebeda - {qinput} {sqinput} - quoted input
			}

			words, err := shellquote.Split(cmdstr)
			if err != nil {
				panic(err)
			}

			cmdname := words[0]
			var cmdParams []string
			for _, word := range words[1:] {
				if strings.Contains(word, " ") && cmdname[0] == '!' {
					cmdParams = append(cmdParams, `"`+word+`"`)
				} else {
					cmdParams = append(cmdParams, word)
				}
			}

			if cmdname[0] == '!' {
				cmdname = strings.Replace(cmdname, "!", "", 1)
			}
			cmd := exec.Command(cmdname, cmdParams...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Start(); err != nil {
				log.Println("Error:", err)
				writeDebugCommand(cmdname, cmdParams)
				fmt.Print("Press 'Enter' to continue...")
				_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			} else {
				if *dbgPtr {
					writeDebugCommand(cmdname, cmdParams)
				} else {
					println("$", cmdstr)
				}

				_, interactive := os.LookupEnv("NU_INTERACTIVE")
				// TODO Lebeda - zajistit nějak jinak, přímo z definice menu aby nečekal i na GUI programy
				// Nápad - do proměnné MENU_INTERACTIVE zapsat seznam vzorů, které se budou spouštět s čekáním -> např. wzt
				// Nápad - do definice menu přidat {wait} a při spouštění jej zlikvidovat
				if interactive {
					if err = cmd.Wait(); err != nil {
						panic(err)
					}
				} else {
					// TODO Lebeda - dořešit jinak prodlevu - pomocí wait?
					time.Sleep(1 * time.Second) // na windows to jinak nefunguje
				}
			}

		}
	}
}

func previewMenu(previewLinePtr *string) {
	prSplit := strings.Split(*previewLinePtr, "\t")

	n := color.New(color.FgWhite)
	h := color.New(color.FgYellow)
	d := color.New(color.FgCyan)
	r := color.New(color.FgGreen)

	h.EnableColor()
	h.Println(prSplit[0])
	h.DisableColor()

	fmt.Println("")

	n.EnableColor()
	n.Print("F1/Enter: ")
	n.DisableColor()
	d.EnableColor()
	d.Println(strings.ReplaceAll(prSplit[1], `\\`, `\`))
	d.DisableColor()

	if len(prSplit) > 2 {
		for i, line := range prSplit[2:] {
			n.EnableColor()
			n.Print("F" + strconv.Itoa(i+2) + ":       ")
			n.DisableColor()
			r.EnableColor()
			r.Println(strings.ReplaceAll(line, `\\`, `\`))
			r.DisableColor()
		}
	}
}

func writeDebugCommand(cmd string, params []string) {
	fmt.Printf("cmd: >%s<\n", cmd)
	for _, w := range params {
		fmt.Printf("param: >%s<\n", w)
	}
}
