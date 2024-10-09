package p2p

import (
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

type GameState struct {
	listenAddr  string
	broadcastch chan BroadcastTo

	currentStatus       *AtomicInt
	currentPlayerAction *AtomicInt
	currentDealer       *AtomicInt
	currentPlayerTurn   *AtomicInt

	// recvPlayerActions *PlayerActionsRecv

	playersList *PlayersList

	table *Table
}

func NewGameState(addr string, bc chan BroadcastTo) *GameState {
	g := &GameState{
		listenAddr:          addr,
		broadcastch:         bc,
		currentStatus:       NewAtomicInt(int32(GameStatusConnected)),
		playersList:         NewPlayersList(),
		currentPlayerAction: NewAtomicInt(1),
		currentDealer:       NewAtomicInt(0),
		// recvPlayerActions:   NewPlayerActionsRecv(),
		currentPlayerTurn: NewAtomicInt(0),
		table:             NewTable(6),
	}

	g.playersList.add(addr)

	go g.loop()

	return g
}

func (g *GameState) canTakeAction(from string) bool {
	currentPlayerAddr := g.playersList.get(g.currentPlayerTurn.Get())
	return currentPlayerAddr == from
}
func (g *GameState) isFromCurrentDealer(from string) bool {
	return g.playersList.get(g.currentDealer.Get()) == from
}

func (g *GameState) handlePlayerAction(from string, action MessagePlayerAction) error {
	if !g.canTakeAction(from) {
		return fmt.Errorf("player (%s) taking action before his turn", from)
	}

	if action.CurrentGameStatus != GameStatus(g.currentStatus.Get()) && !g.isFromCurrentDealer(from) {
		return fmt.Errorf("player (%s) has not correct GameState status (%s)", from, action.CurrentGameStatus)
	}

	g.table.SetPlayerStatus(from, action.CurrentGameStatus)

	// g.recvPlayerActions.addAction(from, action)

	if g.playersList.get(g.currentDealer.Get()) == from {
		g.advanceToNextRound()
	}

	g.incNextPlayer()

	logrus.WithFields(logrus.Fields{
		"we":     g.listenAddr,
		"from":   from,
		"action": action,
	}).Info("recv player action")

	return nil
}

func (g *GameState) TakeAction(action PlayerAction, value int) error {
	if !g.canTakeAction(g.listenAddr) {
		return fmt.Errorf("taking action before its my turn %s", g.listenAddr)
	}

	g.currentPlayerAction.Set(int32(action))
	g.incNextPlayer()

	if g.listenAddr == g.playersList.get(g.currentDealer.Get()) {
		g.advanceToNextRound()
	}

	a := MessagePlayerAction{
		Action:            action,
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
		Value:             value,
	}
	g.sendToPlayers(a, g.getOtherPlayers()...)

	return nil
}

func (g *GameState) getNextGameStatus() GameStatus {
	status := GameStatus(g.currentStatus.Get())
	switch status {
	case GameStatusPreFlop:
		return GameStatusFlop
	case GameStatusFlop:
		return GameStatusTurn
	case GameStatusTurn:
		return GameStatusRiver
	case GameStatusRiver:
		return GameStatusPlayerReady
	default:
		fmt.Printf("invalid status: =>%+v\n", status)
		panic("invalid GameState status")
	}
}

func (g *GameState) advanceToNextRound() {
	// g.recvPlayerActions.clear()
	g.currentPlayerAction.Set(int32(PlayerActionNone))

	if GameStatus(g.currentStatus.Get()) == GameStatusRiver {
		g.TakeSeatAtTable()
		return
	}
	g.currentStatus.Set(int32(g.getNextGameStatus()))
	for i := 0; i < g.playersList.len(); i++ {
		addr := g.playersList.get(i)
		g.table.SetPlayerStatus(addr, GameStatus(g.currentStatus.Get()))
	}
}

func (g *GameState) incNextPlayer() {
	// player, err := g.table.GetPlayerAfter(g.listenAddr)
	// if err != nil {
	// 	panic(err)
	// }

	if g.playersList.len()-1 == int(g.currentPlayerTurn.Get()) {
		g.currentPlayerTurn.Set(0)
		return
	}
	g.currentPlayerTurn.Inc()

	// fmt.Println("the next player on the table is:", player.tablePos)
	// fmt.Println("old wrong value =>", g.currentPlayerTurn)
	// os.Exit(0)
}

func (g *GameState) SetStatus(addr string, s GameStatus) {

	g.table.SetPlayerStatus(addr, s)

	if GameStatus(g.currentStatus.Get()) != s {
		g.table.SetPlayerStatus(g.listenAddr, s)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		g.setStatus(s)
	}

}

func (g *GameState) setStatus(s GameStatus) {
	if s == GameStatusPreFlop {
		g.incNextPlayer()
	}

	if GameStatus(g.currentStatus.Get()) != s {
		g.currentStatus.Set(int32(s))
	}
}

func (g *GameState) getCurrentDealerAddr() (string, bool) {
	currentDealerAddr := g.playersList.get(g.currentDealer.Get())
	return currentDealerAddr, g.listenAddr == currentDealerAddr
}

func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {
	prevPlayer, err := g.table.GetPlayerBefore(g.listenAddr)
	if err != nil {
		panic(nil)
	}
	if from != prevPlayer.addr {
		return fmt.Errorf("received encrypted deck from the wrong player (%s) should be (%s)", from, prevPlayer.addr)
	}

	_, isDealer := g.getCurrentDealerAddr()
	if isDealer && from == prevPlayer.addr {
		g.setStatus(GameStatusPreFlop)
		g.table.SetPlayerStatus(g.listenAddr, GameStatusPreFlop)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		return nil
	}

	dealToPlayer, err := g.table.GetPlayerAfter(g.listenAddr)
	if err != nil {
		panic(nil)
	}

	logrus.WithFields(logrus.Fields{
		"receiveFromPlayer": from,
		"we":                g.listenAddr,
		"dealingToPlayer":   dealToPlayer.addr,
	}).Info("receive cards and going to shuffle")

	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer.addr)
	g.setStatus(GameStatusDealing)

	return nil
}

func (g *GameState) InitiateShuffleAndDeal() {
	fmt.Println("===========================")
	fmt.Println(g.listenAddr)
	fmt.Println("===========================")
	dealToPlayer, err := g.table.GetPlayerAfter(g.listenAddr)
	if err != nil {
		panic(err)
	}

	g.setStatus(GameStatusDealing)
	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer.addr)

	logrus.WithFields(logrus.Fields{
		"we": g.listenAddr,
		"to": dealToPlayer.addr,
	}).Info("dealing cards")
}

func (g *GameState) maybeDeal() {
	if GameStatus(g.currentStatus.Get()) == GameStatusPlayerReady {
		g.InitiateShuffleAndDeal()
	}
}

func (g *GameState) SetPlayerAtTable(addr string) {
	logrus.WithFields(logrus.Fields{
		"we: ":     g.listenAddr,
		"player: ": addr,
	}).Info("setting player status to ready")
	tablePos := g.playersList.getIndex(addr)
	g.table.AddPlayerOnPosition(addr, tablePos)

	if g.table.LenPlayers() < 2 {
		return
	}

	if _, areWeDealer := g.getCurrentDealerAddr(); areWeDealer {
		go func() {
			time.Sleep(time.Second * 8)
			g.maybeDeal()
		}()
	}
}

func (g *GameState) TakeSeatAtTable() {
	tablePos := g.playersList.getIndex(g.listenAddr)
	g.table.AddPlayerOnPosition(g.listenAddr, tablePos)

	// g.sendToPlayers(&proto.TakeSeat{
	// 	Addr: g.listenAddr,
	// }, g.getOtherPlayers()...)
	g.sendToPlayers(MessageReady{}, g.getOtherPlayers()...)
	g.setStatus(GameStatusPlayerReady)
}

func (g *GameState) sendToPlayers(payload any, addr ...string) {
	g.broadcastch <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

	logrus.WithFields(logrus.Fields{
		"payload": payload,
		"players": addr,
		"we":      g.listenAddr,
	}).Info("sending payload to player")
}

func (g *GameState) AddPlayer(from string) {
	g.playersList.add(from)
	sort.Sort(g.playersList)
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)

	for {
		<-ticker.C

		currentDealerAddr, _ := g.getCurrentDealerAddr()
		logrus.WithFields(logrus.Fields{
			"we":             g.listenAddr,
			"playerList":     g.playersList.List(),
			"GameState":      GameStatus(g.currentStatus.Get()),
			"currentDealer":  currentDealerAddr,
			"nextPlayerTurn": g.currentPlayerTurn,
			"action":         PlayerAction(g.currentPlayerAction.Get()),
		}).Info()

		logrus.WithFields(logrus.Fields{
			"we":    g.listenAddr,
			"table": g.table,
		}).Info()

	}
}

func (g *GameState) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList.List() {
		if addr == g.listenAddr {
			continue
		}
		players = append(players, addr)
	}

	return players
}

func (g *GameState) getPositionOnTable() int {
	for i := 0; i < g.playersList.len(); i++ {
		if g.playersList.get(i) == g.listenAddr {
			return i
		}
	}

	panic("player does not exit in the playersList; that should not happen!!!")
}

func (g *GameState) getNextDealer() int {
	panic("TODO")
}
