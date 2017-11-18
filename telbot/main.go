package main

import (
	"flag"
	"log"
)

func main() {
	token := flag.String("token", "", "telegram bot token")
	flag.Parse()

	subh := &SubHandler{}

	mux := NewServeMux()
	mux.Handle("sub", subh)
	if err := mux.NewBotAndServe(*token); err != nil {
		log.Fatalln(err)
	}
}

/*
 * func (hander *MessageHandler) subscribe(chatId int64) error {
 *   msg := botapi.NewMessage(chatId, "subsub")
 *   _, err := hander.bot.Send(msg)
 *   if err != nil {
 *     return fmt.Errorf("could not send msg: %s", err)
 *   }
 *
 *   results := make(chan *accessible.Result)
 *   defer close(results)
 *
 *   go func() {
 *     for r := range results {
 *       bs, err := json.Marshal(r)
 *       if err != nil {
 *         log.Printf("could not marshal result to json: %s", err)
 *         return
 *       }
 *
 *       msg := botapi.NewMessage(chatId, string(bs))
 *       _, err = hander.bot.Send(msg)
 *       if err != nil {
 *         log.Printf("could not send msg: %s", err)
 *         return
 *       }
 *     }
 *   }()
 *
 *   ctx, _ := context.WithCancel(context.Background())
 *   return accessible.Watch(ctx, results, "https://www.baidu.com", 5*time.Second)
 * }
 */
