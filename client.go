package main

import (
	"fmt"
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//GameData represents data received from server
type GameData struct {
	GameID   string
	FieldX   int
	FieldY   int
	Player1  string
	Player2  string
	Status   bool
	P1Field  [][]int
	P2Field  [][]int
	Reserved string
}

type Cell struct {
	Button *widget.Button
	X      int
	Y      int
	Field  string
}

/*Requests to be sent:
GET 	http://sb.mailboxly.info/?gameid=AA22DD5511	-request to get data about game AA22DD5511
GET		http://sb.mailboxly.info/					-(???) request to get IDs of all current games
POST	http://sb.mailboxly.info/					-request to create new game room
PUT		http://sb.mailboxly.info/?gameid=AA22DD5511	-request to edit data in game AA22DD5511
*/

func main() {
	initGUI()
}

func initGUI() {

	window := newWindow()

	newMainContainer(window)

	window.ShowAndRun()
}

//Realizes 'method' request and returns response as []byte type
func sendRequest(method string, uri string) []byte {
	response, err := http.Get(uri)
	if err != nil {
		fmt.Println(err)
	}

	json, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	return json
}

//Initializes new main container, which contains player's nickname (for yet) and 'Start game' button
//which sends all registration info to server to create a new game room
func newMainContainer(window fyne.Window) {
	usernameLabel := widget.NewLabel("Username: ")
	usernameEntry := widget.NewEntry()
	usernameRow := container.NewGridWithRows(2, usernameLabel, usernameEntry)

	//When the button is clicked, it sends POST request to create new game room
	startGameButton := widget.NewButton("Start game", func() {
		//var gameData GameData
		//response := sendRequest("POST", "http://sb.mailboxly.info")
		//err := json.Unmarshal(response, &gameData)
		//if err == nil {
		//	fmt.Println(err)
		//} else {
		//	newGameContainer(window, gameData)
		//}
		newGameContainer(window)
	})
	startGameButton.Resize(fyne.NewSize(150, 50))

	mainContainer := container.NewWithoutLayout(usernameRow, startGameButton)

	mainContainer.Resize(fyne.NewSize(700, 500))

	verticalCenter := mainContainer.Size().Width / 2
	usernameRow.Move(fyne.NewPos(mainContainer.Size().Width/2-usernameRow.Size().Width/2, 50))
	startGameButton.Move(fyne.NewPos(verticalCenter-startGameButton.Size().Width/2, mainContainer.Size().Height-80))

	window.SetContent(mainContainer)
}

//Initializes new game container, which contains game details: both user and bot field,
//'End game' button (for yet), to close current game and open new main container
func newGameContainer(window fyne.Window) {
	//Size of cell fields depends on size of window
	fieldSize := window.Canvas().Size().Width/2 - 75

	userContainer := container.NewAdaptiveGrid(10)
	botContainer := container.NewAdaptiveGrid(10)

	userContainer.Move(fyne.NewPos(50, 50))
	botContainer.Move(fyne.NewPos(fieldSize+100, 50))
	userContainer.Resize(fyne.NewSize(fieldSize, fieldSize))
	botContainer.Resize(fyne.NewSize(fieldSize, fieldSize))

	//Setting cells in both user and bot fields
	setButtons(userContainer, "user")
	setButtons(botContainer, "bot")

	endGameButton := widget.NewButton("End game", func() {
		newMainContainer(window)
	})
	endGameButton.Move(fyne.NewPos(window.Canvas().Size().Width/2-50, window.Canvas().Size().Height-80))
	endGameButton.Resize(fyne.NewSize(100, 50))

	gameContainer := container.NewWithoutLayout(
		userContainer,
		botContainer,
		endGameButton,
	)

	//Adding containers to window
	window.SetContent(gameContainer)
}

//Sets a new window
//(Size should be set in both newWindow() and newMainContainer() methods)
func newWindow() fyne.Window {
	application := app.New()
	window := application.NewWindow("Seabattle")
	window.Resize(fyne.NewSize(700, 500))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	return window
}

//Sets cells on container. Parameter 'field' determines on which field cells should be set
func setButtons(container *fyne.Container, field string) {
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			cell := Cell{
				X:     x,
				Y:     y,
				Field: field,
			}
			cell.Button = widget.NewButton("0", cell.cellTapHandler())

			container.Add(cell.Button)
		}
	}
}

//Handles a click on cell when game stage is active
func (cell Cell) cellTapHandler() func() {
	return func() {
		fmt.Println(cell.Field, ":", cell.X, cell.Y)
	}
}
