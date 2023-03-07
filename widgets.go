package main

import (
	"errors"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"
	dero "github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
)

// / declare some widgets
var (
	primes   = []string{"MAINNET", "TESTNET", "SIMULATOR", "CUSTOM"} /// set select menu
	dropDown = widget.NewSelect(primes, func(s string) {             /// do when select changes
		whichDaemon(s)
		log.Println("[dSlate] Daemon Set To:", s)
	})

	rpcLoginInput  = widget.NewPasswordEntry()
	rpcWalletInput = widget.NewEntry()
	contractInput  = widget.NewMultiLineEntry()

	daemonCheckBox = widget.NewCheck("Daemon Connected", func(value bool) {
		StopGnomon(Gnomes.Init)
	})

	walletCheckBox = widget.NewCheck("Wallet Connected", func(value bool) {
		/// do something on change
	})

	currentHeight = widget.NewEntry()
	walletBalance = widget.NewEntry()

	gnomonEnabled = widget.NewRadioGroup([]string{}, func(s string) {})

	debugEnabled = widget.NewCheck("Debug", func(b bool) {
		if b {
			debug = true
		} else {
			debug = false
		}
	})
)

func rpcLoginEdit() fyne.Widget { /// user:pass password entry
	rpcLoginInput.SetPlaceHolder("RPC user:pass")
	rpcLoginInput.Resize(fyne.NewSize(360, 45))
	rpcLoginInput.Move(fyne.NewPos(10, 650))

	return rpcLoginInput
}

func rpcWalletEdit() fyne.Widget { /// wallet rpc address entry
	rpcWalletInput.SetPlaceHolder("Wallet RPC Address")
	rpcWalletInput.Resize(fyne.NewSize(250, 45))
	rpcWalletInput.Move(fyne.NewPos(10, 700))

	return rpcWalletInput
}

func rpcConnectButton() fyne.Widget { /// wallet connect button
	button := widget.NewButton("Connect", func() { /// do on pressed
		walletAddress = rpcWalletInput.Text
		GetAddress()
	})
	button.Resize(fyne.NewSize(100, 42))
	button.Move(fyne.NewPos(270, 702))

	return button
}

func daemonSelectOption() fyne.Widget { /// daemon select menu
	dropDown.SetSelectedIndex(0)
	dropDown.Resize(fyne.NewSize(180, 45))
	dropDown.Move(fyne.NewPos(10, 550))

	return dropDown
}

func daemonConnectBox() fyne.Widget { /// daemon check box
	daemonCheckBox.Resize(fyne.NewSize(30, 30))
	daemonCheckBox.Move(fyne.NewPos(3, 595))
	daemonCheckBox.Disable()

	return daemonCheckBox
}

func walletConnectBox() fyne.Widget { /// wallet check box
	walletCheckBox.Resize(fyne.NewSize(30, 30))
	walletCheckBox.Move(fyne.NewPos(3, 620))
	walletCheckBox.Disable()

	return walletCheckBox
}

func heightDisplay() fyne.Widget { /// height display entry is read only
	currentHeight.SetText("Height:")
	currentHeight.Disable()
	currentHeight.Resize(fyne.NewSize(170, 45))
	currentHeight.Move(fyne.NewPos(200, 550))

	return currentHeight

}

func balanceDisplay() fyne.Widget {
	walletBalance.SetText("Balance:")
	walletBalance.Disable()
	walletBalance.Resize(fyne.NewSize(170, 45))
	walletBalance.Move(fyne.NewPos(200, 600))

	return walletBalance

}

func contractEdit() fyne.Widget { /// contract entry
	contractInput.SetPlaceHolder("Enter Contract Id:")
	contractInput.Resize(fyne.NewSize(360, 45))
	contractInput.Move(fyne.NewPos(10, 15))
	contractInput.Wrapping = fyne.TextWrapWord

	return contractInput
}

func contractCode() fyne.Widget {
	button := widget.NewButton("SC Code", func() {
		if len(contractInput.Text) == 64 {
			getSCcode(contractInput.Text)
		}
	})

	return button
}

func searchButton() fyne.Widget { /// SC search button
	button := widget.NewButton("Search", func() {
		log.Println("[dSlate] Searching for: " + contractInput.Text)
		p := &dero.GetSC_Params{
			SCID:      contractInput.Text,
			Code:      true,
			Variables: true,
		}
		getSC(p)
	})
	button.Resize(fyne.NewSize(360, 42))
	button.Move(fyne.NewPos(10, 63))
	return button
}

func builtOnImage() fyne.CanvasObject { ///  main image
	img := canvas.NewImageFromResource(resourceBuiltOnDeroPng)
	img.FillMode = canvas.ImageFillOriginal
	img.Resize(fyne.NewSize(380, 540))
	img.Move(fyne.NewPos(10, 210))

	return img
}

func cardImage() fyne.CanvasObject { /// card image
	img := canvas.NewImageFromResource(resourceDero1Png)
	img.FillMode = canvas.ImageFillOriginal
	img.Resize(fyne.NewSize(450, 330))
	img.Move(fyne.NewPos(-33, 200))

	return img
}

func blankWidget() fyne.Widget { /// slate label
	blank := widget.NewLabel("Something goes here...")
	return blank
}

func enableGnomon() fyne.CanvasObject {
	label := widget.NewLabel("Gnomon")
	label.Alignment = fyne.TextAlignCenter
	gnomonEnabled = widget.NewRadioGroup([]string{"On", "Off"}, func(s string) {
		switch s {
		case "On":
			if daemonConnect {
				go startGnomon(daemonAddress)
			} else {
				gnomonEnabled.SetSelected("Off")
			}
		case "Off":
			StopGnomon(Gnomes.Init)
		default:
		}
	})
	gnomonEnabled.Horizontal = true

	cont := container.NewVBox(
		label,
		container.NewCenter(gnomonEnabled))

	return cont
}

func gnomonOpts() fyne.CanvasObject {
	label := widget.NewLabel("")
	label.Wrapping = fyne.TextWrapWord
	kv_entry := widget.NewEntry()
	kv_entry.SetPlaceHolder("Key:")

	korv := widget.NewRadioGroup([]string{"Key", "Value"}, func(s string) {})
	korv.Horizontal = true

	soru := widget.NewRadioGroup([]string{"String", "Uint64"}, func(s string) {})
	soru.Horizontal = true

	search := widget.NewButton("Search", func() {
		if Gnomes.Init {
			switch korv.Selected {
			case "Key":
				switch soru.Selected {
				case "String":
					log.Println("[dSlate] Search results for string key "+kv_entry.Text+" on SCID "+contractInput.Text, searchByKey(contractInput.Text, kv_entry.Text, true))
					label.SetText(searchByKey(contractInput.Text, kv_entry.Text, true))
				case "Uint64":
					log.Println("[dSlate] Search results for uint64 key "+kv_entry.Text+" on SCID "+contractInput.Text, searchByKey(contractInput.Text, kv_entry.Text, false))
					label.SetText(searchByKey(contractInput.Text, kv_entry.Text, false))
				default:
					log.Println("[dSlate] Select string or uint64")
				}
			case "Value":
				switch soru.Selected {
				case "String":
					log.Println("[dSlate] Search results for string value "+kv_entry.Text+" on SCID "+contractInput.Text, searchByValue(contractInput.Text, kv_entry.Text, true))
					label.SetText(searchByValue(contractInput.Text, kv_entry.Text, true))
				case "Uint64":
					log.Println("[dSlate] Search results for uint64 value "+kv_entry.Text+" on SCID "+contractInput.Text, searchByValue(contractInput.Text, kv_entry.Text, false))
					label.SetText(searchByValue(contractInput.Text, kv_entry.Text, false))
				default:
					log.Println("[dSlate] Select string or uint64")
				}
			default:
				log.Println("[dSlate] Select key or value")
			}
		} else {
			log.Println("[dSlate] Gnomon not initialized")
		}

	})

	cont := container.NewVBox(
		label,
		container.NewCenter(korv),
		container.NewCenter(soru),
		container.NewAdaptiveGrid(2, kv_entry, search))

	return cont

}

type nfaAmt struct {
	table.NumericalEntry
}

func nfaOpts() fyne.CanvasObject {
	label := canvas.NewText("", color.White)
	label.TextSize = 18

	file_name := widget.NewEntry()
	file_name.SetPlaceHolder("File Name:")

	start := &nfaAmt{}
	start.ExtendBaseWidget(start)
	start.SetPlaceHolder("Starting at #:")
	start.Validator = validation.NewRegexp(`^\d{1,}`, "Format Not Valid")

	limit := &nfaAmt{}
	limit.ExtendBaseWidget(limit)
	limit.SetPlaceHolder("Ending at #:")
	limit.Validator = validation.NewRegexp(`^\d{1,}`, "Format Not Valid")

	fee := &nfaAmt{}
	fee.ExtendBaseWidget(fee)
	fee.SetPlaceHolder("Fee:")
	fee.Validator = validation.NewRegexp(`^\d{1,}`, "Format Not Valid")

	stop := widget.NewButton("Stop Loop", func() {
		log.Println("[dSlate] Stopping install loop")
		label.Text = "Stopping install loop..."
		label.Refresh()
		kill_process = true
	})

	extension := widget.NewSelect([]string{".jpg", ".png", ".gif", ".mp3", ".mp4", ".pdf", ".zip", ".7z", ".avi", ".mov", ".ogg"}, func(s string) {})
	extension.PlaceHolder = "ext"

	var install fyne.Widget
	install = widget.NewButton("Install Nfas", func() {
		go func() {
			if fee.Validate() == nil && limit.Validate() == nil && start.Validate() == nil {
				if !process_on {
					process_on = true
					install.Hide()
					stop.Show()
					file_name.Disable()
					start.Disable()
					limit.Disable()
					fee.Disable()
					extension.Disable()

					name := file_name.Text
					lim := rpc.StringToInt(limit.Text)
					fe := rpc.StringToInt(fee.Text)
					inc := rpc.StringToInt(start.Text)

					log.Println("[dSlate] Starting install loop for", name+strconv.Itoa(inc)+".bas", "to", name+strconv.Itoa(lim)+".bas")

					for i := 10; i > 0; i-- {
						if kill_process {
							break
						}

						label.Text = "Starting install loop in " + strconv.Itoa(i)
						label.Refresh()
						time.Sleep(1 * time.Second)
					}

					label.Text = ""
					label.Refresh()

					for {
						if kill_process {
							break
						}

						path := name + strconv.Itoa(inc) + ".bas"
						if _, err := os.Stat(path); err == nil {
							log.Println("[dSlate] Installing", path)
							label.Text = "Installing " + path
							label.Refresh()
						} else if errors.Is(err, os.ErrNotExist) {
							log.Println("[dSlate]", path, "Not Found")
							break
						}

						file, err := os.ReadFile(path)

						if err != nil {
							log.Println("[dSlate]", err)
							break
						}

						uploadContract(string(file), uint64(fe))
						inc++

						if inc > lim {
							break
						}

						log.Println("[dSlate] Waiting for block")
						time.Sleep(36 * time.Second)
					}

					label.Text = ""
					label.Refresh()
					install.Show()
					stop.Hide()
					file_name.Enable()
					start.Enable()
					limit.Enable()
					fee.Enable()
					extension.Enable()
					process_on = false
					kill_process = false
					log.Println("[dSlate] Install loop complete")

				} else {
					log.Println("[dSlate] Install already running")
				}
			} else {
				stop.Hide()
				log.Println("[dSlate] Install entries not valid")
			}
		}()
	})

	update := widget.NewButton("Update Contract", func() {
		path := file_name.Text
		if _, err := os.Stat(path); err == nil {
			log.Println("[dSlate] Update Path", path)
			file, err := os.ReadFile(path)

			if err != nil {
				log.Println("[dSlate]", err)
				return
			}
			code := string(file)
			if code != "" {
				fe := rpc.StringToInt(fee.Text)
				updateContract(contractInput.Text, string(file), uint64(fe))
			} else {
				log.Println("[dSlate] Failed to update, code is empty string")
			}

		} else if errors.Is(err, os.ErrNotExist) {
			log.Println("[dSlate]", path, "Not Found")
		}

	})

	stop.Hide()

	wf_entry := widget.NewEntry()
	wf_entry.SetPlaceHolder("Wallet file name:")
	wf_pass := widget.NewPasswordEntry()
	wf_pass.SetPlaceHolder("Wallet file password:")

	verify_button := widget.NewButton("Verify Sign File", func() {
		if wf, err := walletapi.Open_Encrypted_Wallet(wf_entry.Text, wf_pass.Text); err == nil {
			input_file := file_name.Text
			output_file := strings.TrimSuffix(input_file, ".sign")

			if data, err := os.ReadFile(input_file); err != nil {
				log.Println("[dSlate] Cannot read input file", err)
			} else if signer, message, err := wf.CheckSignature(data); err != nil {
				log.Println("[dSlate] Signature verify failed", input_file, err)
			} else {
				log.Println("[dSlate] Signed by", "address", signer.String())

				if os.WriteFile(output_file, message, 0600); err != nil {
					log.Println("[dSlate] Cannot write output file", output_file, err)
				}
				log.Println("[dSlate] Successfully wrote message to file. please check", "file", output_file)
			}

			wf.Close_Encrypted_Wallet()
		} else {
			log.Println("[dSlate] Wallet", err)
		}
	})

	sign_button := widget.NewButton("Sign File", func() {
		if limit.Validate() == nil && start.Validate() == nil {
			if wf, err := walletapi.Open_Encrypted_Wallet(wf_entry.Text, wf_pass.Text); err == nil {
				go func() {
					if !process_on {
						process_on = true
						install.Hide()
						stop.Show()
						file_name.Disable()
						start.Disable()
						limit.Disable()
						fee.Disable()
						extension.Disable()

						ext := extension.Selected
						input := file_name.Text
						lim := rpc.StringToInt(limit.Text)
						inc := rpc.StringToInt(start.Text)

						log.Println("[dSlate] Starting sign loop for", input+strconv.Itoa(inc)+ext, "to", input+strconv.Itoa(lim)+ext)

						for i := 10; i > 0; i-- {
							if kill_process {
								break
							}

							label.Text = "Starting sign loop in " + strconv.Itoa(i)
							label.Refresh()
							time.Sleep(1 * time.Second)
						}

						label.Text = ""
						label.Refresh()

						for {
							if kill_process {
								break
							}

							input_file := input + strconv.Itoa(inc) + ext
							output_file := input_file + ".sign"

							if data, err := os.ReadFile(input_file); err != nil {
								log.Println("[dSlate] Cannot read input file", err)
								break
							} else if err := os.WriteFile(output_file, wf.SignData(data), 0600); err != nil {
								log.Println("[dSlate] Cannot write output file", output_file)
								break
							} else {
								log.Println("[dSlate] Successfully signed file. please check", output_file)
							}

							inc++

							if inc > lim {
								break
							}

							time.Sleep(6 * time.Second)
						}

						label.Text = ""
						label.Refresh()
						install.Show()
						stop.Hide()
						file_name.Enable()
						start.Enable()
						limit.Enable()
						fee.Enable()
						extension.Enable()
						process_on = false
						kill_process = false
						log.Println("[dSlate] Sign loop complete")

					} else {
						log.Println("[dSlate] Loop already running")
					}

					wf.Close_Encrypted_Wallet()
				}()
			} else {
				log.Println("[dSlate] Wallet", err)
			}
		} else {
			log.Println("[dSlate] Sign entries not valid")
		}
	})

	return container.NewVBox(
		container.NewBorder(nil, nil, nil, extension, file_name),
		layout.NewSpacer(),
		verify_button,
		layout.NewSpacer(),
		wf_entry,
		wf_pass,
		sign_button,
		layout.NewSpacer(),
		container.NewCenter(label),
		start,
		limit,
		layout.NewSpacer(),
		fee,
		install,
		stop,
		layout.NewSpacer(),
		update)
}
