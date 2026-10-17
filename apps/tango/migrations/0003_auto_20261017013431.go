package migrations

import "github.com/angvp/tango/migration"

// tango:migration-json [{"app":"shop","name":"0003_auto_20261017013431","up":[{"kind":"CreateIndex","table":"pet","column":"category_id","def":{"Name":"","Type":"","PrimaryKey":false,"Unique":false,"Indexed":false,"References":"","Default":""}},{"kind":"CreateIndex","table":"order","column":"pet_id","def":{"Name":"","Type":"","PrimaryKey":false,"Unique":false,"Indexed":false,"References":"","Default":""}},{"kind":"CreateIndex","table":"order","column":"customer_id","def":{"Name":"","Type":"","PrimaryKey":false,"Unique":false,"Indexed":false,"References":"","Default":""}}],"down":[{"kind":"DropIndex","table":"order","column":"customer_id","def":{"Name":"","Type":"","PrimaryKey":false,"Unique":false,"Indexed":false,"References":"","Default":""}},{"kind":"DropIndex","table":"order","column":"pet_id","def":{"Name":"","Type":"","PrimaryKey":false,"Unique":false,"Indexed":false,"References":"","Default":""}},{"kind":"DropIndex","table":"pet","column":"category_id","def":{"Name":"","Type":"","PrimaryKey":false,"Unique":false,"Indexed":false,"References":"","Default":""}}],"reversible":true}]
var M0003Auto20261017013431 = []migration.Migration{
	{App: "shop", Name: "0003_auto_20261017013431", Reversible: true,
		Up: []migration.Step{
			migration.CreateIndex{Table: "pet", Column: "category_id"},
			migration.CreateIndex{Table: "order", Column: "pet_id"},
			migration.CreateIndex{Table: "order", Column: "customer_id"},
		},
		Down: []migration.Step{
			migration.DropIndex{Table: "order", Column: "customer_id"},
			migration.DropIndex{Table: "order", Column: "pet_id"},
			migration.DropIndex{Table: "pet", Column: "category_id"},
		},
	},
}
