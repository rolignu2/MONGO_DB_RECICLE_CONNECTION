package main

import (
	"core"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"os"
)

func main() {

	//t := core.StrConcat("a", "->", "b")
	//fmt.Println(t)

	mongo := new(core.MongoDB)

	mongo.Conn(true)
	defer mongo.Kill(false)

	//result, err := mongo.FindBy("test", "people", nil, bson.M{"name": "Ale"})
	result, err := mongo.FindBy(nil, bson.M{"name": "Ale"}, "test", "people")

	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR EN en la busqueda")
	} else {
		fmt.Println("entro al resultado pero sin resultados ")
		fmt.Println(result)

	}

	os.Exit(3)
	//session := mongo.GetSession()

}
