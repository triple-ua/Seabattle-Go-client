package main

import (
	"bytes"
	"encoding/json"
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

//Creates and sends new 'method' request to 'uri'. If used to create GET request dataToSend should
//be NIL, because GET request requires no data to send. Method returns not unmarshaled response body,
//so it means to use json.Unmarshal() next to sendRequest()
func sendRequest(method string, uri string, dataToSend []byte) []byte {
	client := &http.Client{}
	var err error
	var request *http.Request
	var response *http.Response

	request, err = http.NewRequest(method, uri, bytes.NewBuffer(dataToSend))
	if err != nil {
		fmt.Println(err)
	}

	//Separate method GET because it is the only method that requires no data to send, but only uri
	if method == "GET" {
		response, err = client.Get(uri)
	} else {
		response, err = client.Do(request)
	}

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
		var gameData GameData
		//response := sendRequest("POST", "http://sb.mailboxly.info")
		response := sendRequest("GET", "http://sb.mailboxly.info/?gameid=AA22DD5511", nil)
		err := json.Unmarshal([]byte(response), &gameData)
		if err != nil {
			fmt.Println(err)
		} else {
			newGameContainer(window, gameData)
		}
	})
	startGameButton.Resize(fyne.NewSize(150, 50))

	mainContainer := container.NewWithoutLayout(usernameRow, startGameButton)

	mainContainer.Resize(fyne.NewSize(700, 500))

	verticalCenter := mainContainer.Size().Width / 2
	usernameRow.Move(fyne.NewPos(mainContainer.Size().Width/2-usernameRow.Size().Width/2, 50))
	startGameButton.Move(fyne.NewPos(verticalCenter-startGameButton.Size().Width/2, mainContainer.Size().Height-80))

	window.SetContent(mainContainer)
	window.SetTitle("Seabattle")
}

//Initializes new game container, which contains game details: both user and bot field,
//'End game' button (for yet), to close current game and open new main container
func newGameContainer(window fyne.Window, gameData GameData) {
	//Size of cell fields depends on size of window
	fieldSize := window.Canvas().Size().Width/2 - 75

	//userContainer := container.NewAdaptiveGrid(10)
	//botContainer := container.NewAdaptiveGrid(10)
	userContainer := container.NewAdaptiveGrid(5)
	botContainer := container.NewAdaptiveGrid(5)

	userContainer.Move(fyne.NewPos(50, 80))
	botContainer.Move(fyne.NewPos(fieldSize+100, 80))
	userContainer.Resize(fyne.NewSize(fieldSize, fieldSize))
	botContainer.Resize(fyne.NewSize(fieldSize, fieldSize))

	player1Label := widget.NewLabel(gameData.Player1 + "'s field:")
	player2Label := widget.NewLabel(gameData.Player2 + "'s field:")
	player1Label.Move(fyne.NewPos(50, 30))
	player2Label.Move(fyne.NewPos(fieldSize+100, 30))

	//Setting cells in both user and bot fields
	userCellArray := setButtons(userContainer, "user")
	botCellArray := setButtons(botContainer, "bot")

	updateCells(userCellArray, gameData.P1Field, userContainer)
	updateCells(botCellArray, gameData.P2Field, botContainer)

	endGameButton := widget.NewButton("End game", func() {
		newMainContainer(window)
	})
	endGameButton.Move(fyne.NewPos(window.Canvas().Size().Width/2-50, window.Canvas().Size().Height-80))
	endGameButton.Resize(fyne.NewSize(100, 50))

	gameContainer := container.NewWithoutLayout(
		player1Label,
		player2Label,
		userContainer,
		botContainer,
		endGameButton,
	)

	//Adding containers to window
	window.SetContent(gameContainer)

	window.SetTitle("Seabattle: Game ID: " + gameData.GameID)
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
func setButtons(container *fyne.Container, fieldName string) [5][5]Cell {
	var cellArray [5][5]Cell = [5][5]Cell{}

	//for x := 0; x < 10; x++ {
	for x := 0; x < 5; x++ {
		//for y := 0; y < 10; y++ {
		for y := 0; y < 5; y++ {
			cell := Cell{
				X:     x,
				Y:     y,
				Field: fieldName,
			}

			cell.Button = widget.NewButton("", func() {
				fmt.Println(
					cell.Field,
					cell.X,
					cell.Y)

				//Marshalling cell's coordinates into dataToSend (type []byte)
				/*dataToSend, err := json.Marshal([]int{cell.X, cell.Y}) //may have problems
				if err != nil {
					fmt.Println(err)
				}
				rawData := sendRequest("PUT", "http://sb.mailboxly.info/?gameid=AA22DD5511", dataToSend)
				//Unmarshal received data from PUT request
				*/
			})

			container.Add(cell.Button)
			cellArray[x][y] = cell
		}
	}

	return cellArray
}

//Updates text in cell.Button and refreshes parent container
func updateCells(cellArray [5][5]Cell, field [][]int, cont *fyne.Container) {
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			cellArray[x][y].Button.Text = fmt.Sprintf("%d", field[x][y])
		}
	}
	cont.Refresh()
}
