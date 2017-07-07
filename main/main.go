package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"time"
)

const (
	ConfPath string = "./src/Config/config.json"
)

type MongoDB struct {
	Mconn  *mgo.Session
	_db    string
	_table string
	_error bool
}

func (this *MongoDB) Catch(err interface{}) {
	log.Println("CONNECTION : [MONGODB] ")
	log.Println(err)
}

func (this *MongoDB) GetConfig() (map[string]interface{}, bool) {

	data, err := ioutil.ReadFile(ConfPath)

	if err != nil {
		this.Catch(err)
		return nil, false
	}

	var resp map[string]interface{}

	if err := json.Unmarshal(data, &resp); err != nil {
		this.Catch(err)
		return nil, false
	}

	return resp, true
}

func (this *MongoDB) PrepareC(Credentials bool) (*mgo.Session, error) {

	conf, Cexist := this.GetConfig()
	var servers string = ""
	var err error
	var session *mgo.Session

	if Cexist == false {
		this.Catch("CONNECTION : [FILE][NO SE ENCUENTRA EL ARCHIVO DE CONFIGURACION] ")
		return nil, errors.New("FILE ERROR")
	}

	if Credentials == false {

		total := len(conf["servers"].([]interface{}))

		for _, value := range conf {
			if rec, ok := value.([]interface{}); ok {
				for key, val := range rec {
					servers += val.(string)
					if key != (total - 1) {
						servers += ","
					}
				}
			}
		}

		if servers == "" {
			this.Catch("CONNECTION: [CONFIG][NO EXISTEN SERVIDORES EN EL ARCHIVO]")
			return nil, errors.New("CONFIG ERROR.")
		}

		session, err = mgo.Dial(servers)

	} else {

		this.Catch("CONNECTION: [FUNCTION][CREANDO CONEXION POR CREDENCIALES ]")

		total := len(conf["servers"].([]interface{}))
		Ostring := make([]string, total)

		for _, value := range conf {
			if rec, ok := value.([]interface{}); ok {
				for _, val := range rec {
					Ostring = append(Ostring, val.(string))
				}
			}

		}

		info := &mgo.DialInfo{
			Addrs:    Ostring,
			Username: conf["user"].(string),
			Password: conf["password"].(string),
		}

		session, err = mgo.DialWithInfo(info)
	}

	t := int64(conf["timeout"].(float64))

	if err != nil {

		defer func() {
			if r := recover(); r == nil {
				this.Catch("ERROR CAUSE : [" + err.Error() + "]")
			}
		}()

	} else {
		session.SetSocketTimeout(time.Hour * (time.Duration(t)))

	}

	return session, err

}

func (this *MongoDB) Conn(Credentials bool) {

	session, err := this.PrepareC(Credentials)

	if err != nil {
		this.Catch("CONNECTION: [ERROR][CONN NO SE PUEDE COMUNICAR]")
		this._error = true
	} else {
		this._error = false
		this.Mconn = session.Clone()
		this.Kill(false, session)
	}

}

func (this *MongoDB) Kill(all bool, session ...*mgo.Session) bool {

	if this._error == true {
		return false
	}

	var sentinel bool = false

	if all == true {

		if session != nil {
			for _, op := range session {
				op.Close()
			}

			sentinel = true
		}

		if this.Mconn != nil {
			this.Mconn.Close()
			sentinel = true
		}

	} else {

		if session != nil {
			for _, op := range session {
				op.Close()
			}

			sentinel = true
		}

	}

	return sentinel
}

func (this *MongoDB) EDatabase(database string, table string) (string, string) {

	var a, b string = "", ""

	if database != "" {
		a = database
	} else {
		a = this._db
	}

	if table != "" {
		b = table
	} else {
		b = this._table
	}

	return a, b
}

func (this *MongoDB) GetSession() *mgo.Session {
	return this.Mconn
}

func (this *MongoDB) InstanceDB(Dbname string) {
	this._db = Dbname
}

func (this *MongoDB) InstanceTable(Table string) {
	this._table = Table
}

func (this *MongoDB) FindBy(pointer interface{}, query bson.M, params ...string) (interface{}, error) {

	if this._error == true {
		return nil, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	err := c.Find(query).One(&pointer)
	return pointer, err
}

func (this *MongoDB) FindAll(pointer interface{}, query bson.M, params ...string) (interface{}, error) {

	if this._error == true {
		return nil, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	e := c.Find(query).All(&pointer)
	return pointer, e
}

func (this *MongoDB) FindAndPaginate(pointer interface{}, query bson.M, start int, end int, params ...string) (interface{}, error) {

	if this._error == true {
		return nil, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	e := c.Find(query).Skip(start).Limit(end).All(&pointer)
	return pointer, e
}

func (this *MongoDB) Count(query bson.M, params ...string) (int, error) {

	if this._error == true {
		return -1, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	r, err := c.Find(query).Count()
	return r, err
}

func (this *MongoDB) CreateCollection() {

}

func main() {

	mongo := new(MongoDB)

	mongo.Conn(true)
	defer mongo.Kill(false)

	//result, err := mongo.FindBy("test", "people", nil, bson.M{"name": "Ale"})
	result, err := mongo.FindBy(nil, bson.M{"name": "Ale"}, "test", "people")

	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR EN en la busqueda")
	} else {
		fmt.Println(result)
	}

	//session := mongo.GetSession()

}
