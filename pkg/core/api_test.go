package core

import (
	"database/sql"
	"errors"
	_ "errors"
	_ "github.com/mattn/go-sqlite3"
	DSN "github.com/tohirov1994/database" //FOR Test Init
	"testing"
)

const dbDriver = "sqlite3"
const dbMemory = ":memory:"

func TestSignIn_QueryError(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_, err = SignIn("", "", db)
	if err == nil {
		t.Errorf("can't execute query: %v", err)
	}

}

func TestInit_CanNotApply(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_ = db.Close()
	err = Init(db)
	if err == nil {
		t.Errorf("Error just not be nil: %v", err)
	}
}

func TestInit_Apply(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	err = Init(db)

	initDDLsDMLs := []string{DSN.ManagersDDL, DSN.ClientsDDL, DSN.ClientsCardsDDL, DSN.AtmsDDL, DSN.ServicesDDL,
		DSN.ManagersDML, DSN.ClientsDML, DSN.ClientsCardsDML, DSN.AtmsDML, DSN.ServicesDML}
	for _, init := range initDDLsDMLs {
		_, err = db.Exec(init)
		if err != nil {
			t.Errorf("can't init db: %v", err)
		}
	}

	if err != nil {
		t.Errorf("init apply, error just be nil: %v", err)
	}
}

func TestSignIn_NoSuchLoginFromEmptyRows(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	_, err = db.Exec(`
	CREATE TABLE managers (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
	login TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL);`)
	if err != nil {
		t.Errorf("can't execute query to base: %v", err)
	}

	result, err := SignIn("", "", db)
	if err != nil {
		t.Errorf("can't execute query from emty rows: %v", err)
	}

	if result != false {
		t.Error("Result signIn no be true, when values account is empty")
	}
}

func TestSignIn_WrongPassword(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	_, err = db.Exec(`
	CREATE TABLE managers (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
	login TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL);`)
	if err != nil {
		t.Errorf("can't execute query to base: %v", err)
	}

	_, err = db.Exec(`INSERT INTO managers(login, password) VALUES ('jack', 'password');`)
	if err != nil {
		t.Errorf("can't execute insert login and password to DB: %v", err)
	}

	_, err = SignIn("jack", "12345", db)
	if !errors.Is(err, PassWrong) {
		t.Errorf("Not PassWrong error for invalid pass: %v", err)
	}
}

func TestSignIn_SignInOk(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	_, err = db.Exec(`
	CREATE TABLE managers (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
	login TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL);`)
	if err != nil {
		t.Errorf("can't execute query to base: %v", err)
	}

	_, err = db.Exec(`INSERT INTO managers(login, password) VALUES ('jack', 'password'),
('max', 'password2');`)
	if err != nil {
		t.Errorf("can't execute insert login and password to DB: %v", err)
	}

	result, err := SignIn("max", "password2", db)
	if err != nil {
		t.Errorf("can't execute signIn: %v", err)
	}
	if result != true {
		t.Error("signIn result must be true for existing account")
	}
}

func TestAddClient_NotPingToDB(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_ = db.Close()
	err = AddClient(`1`, `2`, `3`, `4`, db)
	if err == nil {
		t.Errorf("can't add client: %v", err)
	}
}

func TestAddClient_WithoutTable(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	err = AddClient(`Jack`, `Jackson`, `login`, `pass`, db)
	if err == nil {
		t.Errorf("error just not been nil: %v", err)
	}
}

func TestAddClient_ClientApply(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS clients
(
    Id       INTEGER PRIMARY KEY AUTOINCREMENT,
    name     TEXT    NOT NULL,
    surname  TEXT    NOT NULL,
    login    TEXT    NOT NULL UNIQUE,
    password TEXT    NOT NULL
);`)
	if err != nil {
		t.Errorf("can't execute query to base: %v", err)
	}
	err = AddClient(`Jack`, `Jackson`, `login`, `pass`, db)
	if err != nil {
		t.Errorf("error just be nil: %v", err)
	}
}

func TestAddCardToClient_WithoutCardsTable(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	err = AddCardToClient(1234, 4444, 1000000, "Jack Jackson", 333, 1222,1, db)
	if err == nil {
		t.Errorf("error just not been nil: %v", err)
	}
}

func TestAddCardToClient_AddCardCanceled(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_ = db.Close()
	err = AddCardToClient(1234, 4444, 1000000, `Jack Jackson`, 333, 1222, 1, db)
	if err == nil {
		t.Errorf("We have just be error: %v", err)
	}
}

func TestAddCardToClient_CardApply(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS clients_cards
(
    Id         INTEGER PRIMARY KEY AUTOINCREMENT,
    pan        INTEGER NOT NULL UNIQUE,
    pin        INTEGER NOT NULL,
    balance    INTEGER NOT NULL,
    holderName TEXT    NOT NULL,
    cvv        INTEGER NOT NULL,
    validity   INTEGER NOT NULL,
    client_id  INTEGER NOT NULL REFERENCES clients
);`)
	if err != nil {
		t.Errorf("can't execute query to base: %v", err)
	}
	err = AddCardToClient(1234, 4444, 1000000, `Jack Jackson`, 333, 1222, 1, db)
	if err != nil {
		t.Errorf("error just be nil: %v", err)
	}
}

func TestAddServiceToTheBank_WithoutTable(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	err = AddServiceToTheBank("JackDenial", db)
	if err == nil {
		t.Errorf("error just not been nil: %v", err)
	}
}

func TestAddServiceToTheBank_AddServiceCanceled(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_ = db.Close()
	err = AddServiceToTheBank("Water", db)
	if err == nil {
		t.Errorf("Erroo, just be error: %v", err)
	}
}

func TestAddServiceToTheBank_ApplyService(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS services
(
    Id      INTEGER PRIMARY KEY AUTOINCREMENT,
    service TEXT    NOT NULL,
    balance INTEGER
);`)
	if err != nil {
		t.Errorf("can't execute query to base: %v", err)
	}
	err = AddServiceToTheBank("JackDenial", db)
	if err != nil {
		t.Errorf("error just be nil: %v", err)
	}
}

func TestAddAtmToTheBank_WithoutExistATMsTable(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	err = AddAtmToTheBank("NewYork", "Manhattan", "7 Av/W 47 St", db)
	if err == nil {
		t.Errorf("error just not been nil: %v", err)
	}
}

func TestAddServiceToTheBank_TransactionATMCanceled(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_ = db.Close()
	err = AddAtmToTheBank("New York", "Manhattan", "7 Av/W 47S St", db)
	if err == nil {
		t.Errorf("can't add ATM to bank: %v", err)
	}
}

func TestAddServiceToTheBank_TransactionATMApply(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't opne db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS atms
(
 --   Id       INTEGER PRIMARY KEY AUTOINCREMENT,
    City     TEXT NOT NULL,
    District TEXT NOT NULL,
    Street   TEXT NOT NULL
);`)
	if err != nil {
		t.Errorf("can't create table: %v", err)
	}
	err = AddAtmToTheBank("New York", "Manhattan", "7 Av/W 47S St", db)
	if err != nil {
		t.Errorf("can't add Atm: %v", err)
	}

}

func TestPANLastPlusOne_QueryError(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't opne db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	pan, err := PANLastPlusOne(db)
	if err == nil {
		t.Errorf("We just have Error: %v", err)
	}
	if pan != 0 {
		t.Errorf("just be zero: %d", pan)
	}
}

/*
func TestPANLastPlusOne_QueryOK(t *testing.T) {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't opne db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_, err = db.Query(`
CREATE TABLE clients_cards
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    pan        INTEGER NOT NULL UNIQUE,
    pin        INTEGER NOT NULL,
    balance    INTEGER NOT NULL,
    holderName TEXT    NOT NULL,
    cvv        INTEGER NOT NULL,
    validity   INTEGER NOT NULL,
    client_id  INTEGER NOT NULL REFERENCES clients
);`)
	if err != nil {
		t.Errorf("can't create table for lastPAN %v", err)
	}
	_, err = db.Query(`
INSERT INTO clients_cards VALUES 
(1, 2021600000000000, 1994, 1000000, 'ADMIN CLIENT', 333, 0222, 1);
`)
	if err != nil {
		t.Errorf("can't insert data to table for lastPAN %v", err)
	}
	pan, err := PANLastPlusOne(db)
	if err != nil {
		t.Errorf("just be nil: %v", err)
	}
	if pan != 2021600000000001 {
		t.Errorf("just be 256: %d", pan)
	}
}
*/

func TestCheckIdClient_Ok(t *testing.T)  {
	db, err := sql.Open(dbDriver, dbMemory)
	if err != nil {
		t.Errorf("can't opne db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()
	_, err = db.Query(`
CREATE TABLE clients_cards
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    pan        INTEGER NOT NULL UNIQUE,
    pin        INTEGER NOT NULL,
    balance    INTEGER NOT NULL,
    holderName TEXT    NOT NULL,
    cvv        INTEGER NOT NULL,
    validity   INTEGER NOT NULL,
    client_id  INTEGER NOT NULL REFERENCES clients
);`)
	if err != nil {
		t.Errorf("can't create table for lastPAN %v", err)
	}
	_, err = db.Query(`
INSERT INTO clients_cards VALUES 
(1, 2021600000000000, 1994, 1000000, 'ADMIN CLIENT', 333, 0222, 1);
`)
	if err != nil {
		t.Errorf("can't insert data to table for lastPAN %v", err)
	}
}