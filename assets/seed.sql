CREATE TABLE if not exists "resource" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE,
	"generator"	TEXT,
	"template"	TEXT,
  CONSTRAINT name_generator_template UNIQUE(name, generator, template) ON CONFLICT IGNORE,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "locale" (
	"id" INTEGER NOT NULL UNIQUE,
	"name" TEXT NOT NULL UNIQUE,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "person_prop" (
	"id"	INTEGER NOT NULL UNIQUE,
	"locale_id" INTEGER NOT NULL,
	"type"	TEXT NOT NULL,
	"value"	TEXT,
  CONSTRAINT locale_type_value UNIQUE(locale_id, type, value) ON CONFLICT IGNORE,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "misc_prop" (
	"id"	INTEGER NOT NULL UNIQUE,
	"locale_id" INTEGER NOT NULL,
	"type"	TEXT NOT NULL,
	"value"	TEXT,
  CONSTRAINT locale_type_value UNIQUE(locale_id, type, value) ON CONFLICT IGNORE,
	PRIMARY KEY("id" AUTOINCREMENT)
);

insert or ignore into locale values 
	(null, "fr-FR"),
	(null, "en-UK"),
	(null, "en-US"),
	(null, "es-ES")
	;

insert or ignore into `resource` values 
	(null, "person.firstName", "random_row", "person_prop:type=firstName"),
	(null, "person.lastName", "random_row", "person_prop:type=lastName"),
	(null, "person.nickName", "random_row", "person_prop:type=nickName"),
	(null, "person.age", "union", "person.age.baby|person.age.child|person.age.teen|person.age.adult|person.age.mid|person.age.old"),
	(null, "person.age.baby", "int_range", "1..3"),
	(null, "person.age.child", "int_range", "3..12"),
	(null, "person.age.teen", "int_range", "12..16"),
	(null, "person.age.adult", "int_range", "16..30"),
	(null, "person.age.mid", "int_range", "30..55"),
	(null, "person.age.old", "int_range", "55..100"),
	(null, "person.phone", "union", "person.phone.mobile|person.phone.land"),
	(null, "person.phone.mobile", "pattern", "+33 [6..7] [00..99] [00..99] [00..99] [00..99]"),
	(null, "person.phone.land", "pattern", "+33 [1..9] [00..99] [00..99] [00..99] [00..99]"),
	(null, "person.email", "email", "{firstName}.{lastName}@{provider},{lastName}.{firstName}@{provider},{nickName}@provider"),
	(null, "location.country", "random_row", "location_prop:type=country"),
	(null, "location.town", "random_row", "location_prop:type=town"),
	(null, "location.continent", "random_row", "location_prop:type=continent"),
	(null, "misc.health-insurance", "random_row", "misc_prop:type=health-insurance")
    ;
	

insert or replace into person_prop values 
  (null, 1, "nickName", "le puant"),
  (null, 1, "nickName", "le beau"),
  (null, 1, "nickName", "la peche")
  ;