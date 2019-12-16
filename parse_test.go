package generator

import "testing"

func TestCompileQuery(t *testing.T) {
	table := []struct {
		Q, R, D, T, N string
		V             []string
	}{
		// basic test for named parameters, invalid char ',' terminating
		{
			Q: `INSERT INTO foo (a,b,c,d) VALUES (:name, :age, :first, :last)`,
			R: `INSERT INTO foo (a,b,c,d) VALUES (?, ?, ?, ?)`,
			D: `INSERT INTO foo (a,b,c,d) VALUES ($1, $2, $3, $4)`,
			T: `INSERT INTO foo (a,b,c,d) VALUES (@p1, @p2, @p3, @p4)`,
			N: `INSERT INTO foo (a,b,c,d) VALUES (:name, :age, :first, :last)`,
			V: []string{"name", "age", "first", "last"},
		},
		// This query tests a named parameter ending the string as well as numbers
		{
			Q: `SELECT * FROM a WHERE first_name=:name1 AND last_name=:name2`,
			R: `SELECT * FROM a WHERE first_name=? AND last_name=?`,
			D: `SELECT * FROM a WHERE first_name=$1 AND last_name=$2`,
			T: `SELECT * FROM a WHERE first_name=@p1 AND last_name=@p2`,
			N: `SELECT * FROM a WHERE first_name=:name1 AND last_name=:name2`,
			V: []string{"name1", "name2"},
		},
		{
			Q: `SELECT ":foo" FROM a WHERE first_name=:name1 AND last_name=:name2`,
			R: `SELECT ":foo" FROM a WHERE first_name=? AND last_name=?`,
			D: `SELECT ":foo" FROM a WHERE first_name=$1 AND last_name=$2`,
			T: `SELECT ":foo" FROM a WHERE first_name=@p1 AND last_name=@p2`,
			N: `SELECT ":foo" FROM a WHERE first_name=:name1 AND last_name=:name2`,
			V: []string{"name1", "name2"},
		},
		{
			Q: `SELECT 'a:b:c' || first_name, ':ABC:_:' FROM person WHERE first_name=:first_name AND last_name=:last_name`,
			R: `SELECT 'a:b:c' || first_name, ':ABC:_:' FROM person WHERE first_name=? AND last_name=?`,
			D: `SELECT 'a:b:c' || first_name, ':ABC:_:' FROM person WHERE first_name=$1 AND last_name=$2`,
			T: `SELECT 'a:b:c' || first_name, ':ABC:_:' FROM person WHERE first_name=@p1 AND last_name=@p2`,
			N: `SELECT 'a:b:c' || first_name, ':ABC:_:' FROM person WHERE first_name=:first_name AND last_name=:last_name`,
			V: []string{"first_name", "last_name"},
		},
		{
			Q: `SELECT @name := "name", :age, :first, :last`,
			R: `SELECT @name := "name", ?, ?, ?`,
			D: `SELECT @name := "name", $1, $2, $3`,
			N: `SELECT @name := "name", :age, :first, :last`,
			T: `SELECT @name := "name", @p1, @p2, @p3`,
			V: []string{"age", "first", "last"},
		},
		{
			Q: `INSERT INTO foo (a,b,c,d) VALUES (:あ, :b, :キコ, :名前)`,
			R: `INSERT INTO foo (a,b,c,d) VALUES (?, ?, ?, ?)`,
			D: `INSERT INTO foo (a,b,c,d) VALUES ($1, $2, $3, $4)`,
			N: `INSERT INTO foo (a,b,c,d) VALUES (:あ, :b, :キコ, :名前)`,
			T: `INSERT INTO foo (a,b,c,d) VALUES (@p1, @p2, @p3, @p4)`,
			V: []string{"あ", "b", "キコ", "名前"},
		},
		{
			Q: "-- A Line Comment should be ignored for :params\nINSERT INTO foo (a,b,c,d) VALUES (:あ, :b, :キコ, :名前)",
			R: "-- A Line Comment should be ignored for :params\nINSERT INTO foo (a,b,c,d) VALUES (?, ?, ?, ?)",
			D: "-- A Line Comment should be ignored for :params\nINSERT INTO foo (a,b,c,d) VALUES ($1, $2, $3, $4)",
			N: "-- A Line Comment should be ignored for :params\nINSERT INTO foo (a,b,c,d) VALUES (:あ, :b, :キコ, :名前)",
			T: "-- A Line Comment should be ignored for :params\nINSERT INTO foo (a,b,c,d) VALUES (@p1, @p2, @p3, @p4)",
			V: []string{"あ", "b", "キコ", "名前"},
		},
		{
			Q: `/* A Block Comment should be ignored for :params */INSERT INTO foo (a,b,c,d) VALUES (:あ, :b, :キコ, :名前)`,
			R: `/* A Block Comment should be ignored for :params */INSERT INTO foo (a,b,c,d) VALUES (?, ?, ?, ?)`,
			D: `/* A Block Comment should be ignored for :params */INSERT INTO foo (a,b,c,d) VALUES ($1, $2, $3, $4)`,
			N: `/* A Block Comment should be ignored for :params */INSERT INTO foo (a,b,c,d) VALUES (:あ, :b, :キコ, :名前)`,
			T: `/* A Block Comment should be ignored for :params */INSERT INTO foo (a,b,c,d) VALUES (@p1, @p2, @p3, @p4)`,
			V: []string{"あ", "b", "キコ", "名前"},
		},
		// Repeated names are not distinct in the names list
		{
			Q: `INSERT INTO foo (a,b,c,d) VALUES (:name, :age, :name)`,
			R: `INSERT INTO foo (a,b,c,d) VALUES (?, ?, ?)`,
			D: `INSERT INTO foo (a,b,c,d) VALUES ($1, $2, $3)`,
			T: `INSERT INTO foo (a,b,c,d) VALUES (@p1, @p2, @p3)`,
			N: `INSERT INTO foo (a,b,c,d) VALUES (:name, :age, :name)`,
			V: []string{"name", "age", "name"},
		},
	}

	for _, test := range table {
		qr, names, err := Query([]byte(test.Q), QUESTION, false)
		if err != nil {
			t.Error(err)
		}
		if qr != test.R {
			t.Errorf("expected %s, got %s", test.R, qr)
		}
		if len(names) != len(test.V) {
			t.Errorf("expected %#v, got %#v", test.V, names)
		} else {
			for i, name := range names {
				if name != test.V[i] {
					t.Errorf("expected %dth name to be %s, got %s", i+1, test.V[i], name)
				}
			}
		}

		qd, _, _ := Query([]byte(test.Q), DOLLAR, false)
		if qd != test.D {
			t.Errorf("\nexpected: `%s`\ngot:      `%s`", test.D, qd)
		}

		qt, _, _ := Query([]byte(test.Q), AT, false)
		if qt != test.T {
			t.Errorf("\nexpected: `%s`\ngot:      `%s`", test.T, qt)
		}

		qq, _, _ := Query([]byte(test.Q), NAMED, false)
		if qq != test.N {
			t.Errorf("\nexpected: `%s`\ngot:      `%s`\n(len: %d vs %d)", test.N, qq, len(test.N), len(qq))
		}
	}
}

func TestNamedQueryWithoutParams(t *testing.T) {
	var queries = []string{
		// Array Slice Syntax
		`SELECT schedule[1:2][1:1] FROM sal_emp WHERE name = 'Bill';`,
		`SELECT f1[1][-2][3] AS e1, f1[1][-1][5] AS e2 FROM (SELECT '[1:1][-2:-1][3:5]={{{1,2,3},{4,5,6}}}'::int[] AS f1) AS ss;`,
		`SELECT array_dims(1 || '[0:1]={2,3}'::int[]);`,
		// String Constant Syntax
		`'Dianne'':not_a_parameter horse'`,
		`'Dianne'''':not_a_parameter horse'`,
		`SELECT ':not_an_parameter'`,
		`$$Dia:not_an_parameter's horse$$`,
		`$$Dianne's horse$$`,
		`SELECT 'foo'
			'bar';`,
		`E'user\'s log'`,
		`$$escape ' with ''$$`,
		// Quoted Ident Syntax
		`SELECT "addr:city" FROM "location";`,
		// Type Cast Syntax
		`select '1'   ::   numeric;`, `select '1'   ::  text :: numeric;`,
		// Nested Block Quotes
		`SELECT * FROM users
		/* Ignore all things who aren't after a certain :date
		 * More lines /* nested block comment
		 */*/
		WHERE some_text LIKE 'foo -- bar'`,
	}

	for _, q := range queries {
		qr, names, err := Query([]byte(q), QUESTION, false)
		if err != nil {
			t.Error(err)
		}
		if qr != q {
			t.Errorf("expected query to be unaltered\nexpected: %s\ngot:%s", q, qr)
		}
		if len(names) > 0 {
			t.Errorf("expected params to be empty got: %v", names)
		}
	}
}
