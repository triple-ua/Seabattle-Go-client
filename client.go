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
	Player1  string
	Player2  string
	Status   bool
	P1Field  [][]int
	P2Field  [][]int
	TurnFlag int
}

type Cell struct {
	Button *widget.Button
	X      int
	Y      int
}

type NumberOfShips struct {
	SingleDeck int
	DoubleDeck int
	ThreeDeck  int
	FourDeck   int
	TotalDecks int
}

type Ship struct {
	Size             int
	Orientation      string
	BaseDeckPosition [2]int //position of top/left corner of ship
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
	var gameData GameData
	var shipsArray NumberOfShips

	usernameLabel := widget.NewLabel("Username: ")
	usernameEntry := widget.NewEntry()
	usernameRow := container.NewGridWithRows(2, usernameLabel, usernameEntry)

	userContainer := container.NewAdaptiveGrid(10)
	userContainer.Resize(fyne.NewSize(250, 250))

	shipsRadioGroup := widget.NewRadioGroup(
		[]string{"Single-deck ship", "Double-deck ship", "Three-deck ship", "Four-deck ship"},
		func(s string) {})
	shipsOrientation := widget.NewRadioGroup(
		[]string{"Horizontal", "Vertical"}, func(s string) {})
	shipsRadioGroup.SetSelected("Four-deck ship")
	shipsOrientation.SetSelected("Vertical")
	shipsContainer := container.NewVBox(shipsRadioGroup, widget.NewSeparator(), shipsOrientation)

	userCellArray := setButtons(userContainer, "putShip", [10][10]string{})

	//When the button is clicked, it sends POST request to create new game room
	startGameButton := widget.NewButton("Start game", func() {
		if shipsArray.TotalDecks == 20 {
			//response := sendRequest("GET", "http://localhost:8080/?gameid=AA22DD5511", nil)
			response := sendRequest("GET", "http://localhost:8080/game", nil)
			err := json.Unmarshal([]byte(response), &gameData)

			if err != nil {
				fmt.Println(err)
			} else {
				var buttonValuesArray [10][10]string
				for x := 0; x < 10; x++ {
					for y := 0; y < 10; y++ {
						buttonValuesArray[x][y] = userCellArray[x][y].Button.Text
					}
				}

				newGameContainer(window, gameData, buttonValuesArray)
			}
		} else {
			fmt.Println("Your ships set uncorrectly")
		}
	})
	startGameButton.Resize(fyne.NewSize(150, 50))

	mainContainer := container.NewWithoutLayout(usernameRow, userContainer, startGameButton, shipsContainer)
	mainContainer.Resize(fyne.NewSize(700, 500))

	verticalCenter := mainContainer.Size().Width / 2
	usernameRow.Move(fyne.NewPos(verticalCenter-usernameRow.Size().Width/2, 30))
	userContainer.Move(fyne.NewPos(verticalCenter-userContainer.Size().Width/2, 130))
	startGameButton.Move(fyne.NewPos(verticalCenter-startGameButton.Size().Width/2, mainContainer.Size().Height-100))
	shipsContainer.Move(fyne.NewPos(30, userContainer.Position().Y))

	window.SetContent(mainContainer)
	window.SetTitle("Sea Battle")
}

//Initializes new game container, which contains game details: both user and bot field,
//'End game' button (for yet), to close current game and open new main container
func newGameContainer(window fyne.Window, gameData GameData, buttonValuesArray [10][10]string) {
	//Size of cell fields depends on size of window
	fieldSize := window.Canvas().Size().Width/2 - 75

	userContainer := container.NewAdaptiveGrid(10)
	botContainer := container.NewAdaptiveGrid(10)

	userContainer.Move(fyne.NewPos(50, 80))
	botContainer.Move(fyne.NewPos(fieldSize+100, 80))
	userContainer.Resize(fyne.NewSize(fieldSize, fieldSize))
	botContainer.Resize(fyne.NewSize(fieldSize, fieldSize))

	player1Label := widget.NewLabel(gameData.Player1 + "'s field:")
	player2Label := widget.NewLabel(gameData.Player2 + "'s field:")
	player1Label.Move(fyne.NewPos(50, 30))
	player2Label.Move(fyne.NewPos(fieldSize+100, 30))

	//Setting cells in fields
	userCellArray := setButtons(userContainer, "", buttonValuesArray)
	botCellArray := setButtons(botContainer, "shoot", [10][10]string{})
	fmt.Sprintf("%s", botCellArray, userCellArray)

	endGameButton := widget.NewButton("End game", func() {
		newMainContainer(window)
	})
	endGameButton.Move(fyne.NewPos(window.Canvas().Size().Width/2-50, window.Canvas().Size().Height-100))
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

	window.SetTitle("Sea Battle: Game ID: " + gameData.GameID)
}

//Sets a new window
//(Size should be set in both newWindow() and newMainContainer() methods)
func newWindow() fyne.Window {
	application := app.New()
	window := application.NewWindow("")
	window.Resize(fyne.NewSize(700, 500))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	return window
}

//Sets cells on container. Parameter 'field' determines on which field cells should be set
func setButtons(container *fyne.Container, listener string, initTextValues [10][10]string) [10][10]Cell {
	var cellArray [10][10]Cell = [10][10]Cell{}

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			cell := Cell{
				X: x,
				Y: y,
			}

			cell.Button = widget.NewButton(initTextValues[x][y], func() {
				switch listener {
				case "putShip":
					cell.putShip(cellArray)
					container.Refresh()
				case "shoot":
					//PUT request
				}
			})

			/*if cell.ParentName == "user" {
				cell.putShip(cellArray)
			} else if cell.ParentName == "bot" {
				fmt.Println("PUT request here")
			}*/

			/*//Marshalling cell's coordinates into dataToSend (type []byte)
			dataToSend, err := json.Marshal([]int{cell.X, cell.Y}) //may have problems
			if err != nil {
				fmt.Println(err)
			}
			rawData := sendRequest("PUT", "http://sb.mailboxly.info/?gameid=AA22DD5511", dataToSend)
			//Unmarshal received data from PUT request*/

			container.Add(cell.Button)
			cellArray[x][y] = cell
		}
	}

	return cellArray
}

//Puts a piece of ship in pressed cell if this cell is valid
func (cell Cell) putShip(cellArray [10][10]Cell) {
	//clear if pressed button have "O" text
	if cell.Button.Text == "O" {
		cell.Button.Text = ""
		return

		//check if corners around pressed cell are clear
		//if branch - corners aren't clear; terminate method
		//else branch - corners are clear; next validation stage
	} else if !cell.checkIfCornersAreClear(cellArray) {
		return

	} else {
		drawShip(newShip("Horizontal", "Three-deck ship", cell), &cellArray)
	}
}

//Draws a ship in cell array
func drawShip(ship Ship, cellArray *[10][10]Cell) {
	x := ship.BaseDeckPosition[0]
	y := ship.BaseDeckPosition[1]

	for i := 0; i < ship.Size; i++ {
		switch ship.Orientation {
		case "Horizontal":
			cellArray[x][y+i].Button.Text = "O"
		case "Vertical":
			cellArray[x+i][y].Button.Text = "O"
		}
	}
}

//Creates new ship with given parameter
func newShip(orientation string, size string, cell Cell) Ship {
	var ship Ship

	switch size {
	case "Single-deck ship":
		ship.Size = 1
	case "Double-deck ship":
		ship.Size = 2
	case "Three-deck ship":
		ship.Size = 3
	case "Four-deck ship":
		ship.Size = 4
	}

	ship.BaseDeckPosition = [2]int{cell.X, cell.Y}
	ship.Orientation = orientation

	return ship
}

//Validation method. Returns true if corners are clear, and false if corners aren't
func (cell Cell) checkIfCornersAreClear(cellArray [10][10]Cell) bool {
	for x := -1; x <= 1; x++ {

		//skipping x==0, because method requires only -1 and 1 values
		if x == 0 {
			continue
		}

		for y := -1; y <= 1; y++ {

			//skipping y==0, because method requires only -1 and 1 values
			if y == 0 {
				continue
			}

			//going to next iteration if cell.X is 0 or 9
			//to get rid of index out of range exception
			if cell.X == 0 && x == -1 || cell.X == 9 && x == 1 {
				break
			}

			//going to next iteration if cell.Y is 0 or 9
			//to get rid of index out of range exception
			if cell.Y == 0 && y == -1 || cell.Y == 9 && y == 1 {
				continue
			}

			//false returns if atleast one of corners around cell have "O" text
			if cellArray[cell.X+x][cell.Y+y].Button.Text == "O" {
				return false
			}
		}
	}

	return true
}
