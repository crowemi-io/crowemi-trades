package dater

func main() {
	// ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer stop()

	// c, err := config.Bootstrap(os.Getenv("CONFIG_PATH"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// c.Logger.Log("msg", "start crowemi-trades.dater")

	// var n *notifier.Notifier = nil
	// if c.Notifier.Telegram != nil && c.Notifier.Telegram.BotToken != "" && c.Notifier.Telegram.ChatID != 0 {
	// 	n, err = notifier.New(notifier.Config{
	// 		BotToken: c.Notifier.Telegram.BotToken,
	// 		ChatID:   c.Notifier.Telegram.ChatID,
	// 	})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// // stream watcher init
	// var wat *stream.Watcher = nil
	// // get the symbols associated with app category
	// docs, err := firestoreDB.Client.Doc("accounts/" + c.Alpaca.AccountID).Collection(db.CollectionAllocations).Doc("app").Collection("symbols").Documents(context.TODO()).GetAll()
	// if err != nil {
	// 	c.Logger.Log("msg", "failed to load symbols for watcher", "err", err, "AccountID", c.Alpaca.AccountID)
	// }

	// if docs != nil {
	// 	var symbols []string = nil
	// 	for _, doc := range docs {
	// 		var symbol models.Symbol
	// 		doc.DataTo(symbol)
	// 		symbols = append(symbols, symbol.ID)
	// 	}
	// 	if len(symbols) > 0 {
	// 		wat = &stream.Watcher{
	// 			Logger:         c.Logger,
	// 			Symbols:        symbols,
	// 			APIKey:         c.Alpaca.APIKey,
	// 			APISecret:      c.Alpaca.APISecretKey,
	// 			DataURL:        c.Alpaca.APIDataURL,
	// 			MarketDataFeed: c.Alpaca.MarketDataFeed,
	// 		}
	// 	}
	// } else if err != nil {
	// 	c.Logger.Log("msg", "failed to load portfolio for minute bars", "err", err, "AccountID", c.Alpaca.AccountID)
	// }

	// wat.Run(ctx)

}
