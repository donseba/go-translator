package translator

// Code generated by script. DO NOT EDIT.

var LanguageHeaderTemplates = map[string]TranslationFileHeader{
	"mt": {PluralForms: "nplurals=5; plural=(n == 1 ? 0 : n == 2 ? 1 : n == 0 || n % 100 >= 3 && n % 100 <= 10 ? 2 : n % 100 >= 11 && n % 100 <= 19 ? 3 : 4);"},
	"pt_PT": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"shi": {PluralForms: "nplurals=3; plural=(n == 0 || n == 1 ? 0 : n >= 2 && n <= 10 ? 1 : 2);"},
	"smj": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"tpi": {PluralForms: "nplurals=1; plural=(0);"},
	"doi": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ff": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"kw": {PluralForms: "nplurals=6; plural=(n == 0 ? 0 : n == 1 ? 1 : (n % 100 == 2 || n % 100 == 22 || n % 100 == 42 || n % 100 == 62 || n % 100 == 82) || n % 1000 == 0 && (n % 100000 >= 1000 && n % 100000 <= 20000 || n % 100000 == 40000 || n % 100000 == 60000 || n % 100000 == 80000) || n != 0 && n % 1000000 == 100000 ? 2 : (n % 100 == 3 || n % 100 == 23 || n % 100 == 43 || n % 100 == 63 || n % 100 == 83) ? 3 : n != 1 && (n % 100 == 1 || n % 100 == 21 || n % 100 == 41 || n % 100 == 61 || n % 100 == 81) ? 4 : 5);"},
	"nqo": {PluralForms: "nplurals=1; plural=(0);"},
	"sw": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"tk": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"yue": {PluralForms: "nplurals=1; plural=(0);"},
	"bs": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"gv": {PluralForms: "nplurals=4; plural=(n % 10 == 1 ? 0 : n % 10 == 2 ? 1 : (n % 100 == 0 || n % 100 == 20 || n % 100 == 40 || n % 100 == 60 || n % 100 == 80) ? 2 : 3);"},
	"lld": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"naq": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"sg": {PluralForms: "nplurals=1; plural=(0);"},
	"ug": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"zh": {PluralForms: "nplurals=1; plural=(0);"},
	"bez": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"lag": {PluralForms: "nplurals=3; plural=(n == 0 ? 0 : n == 1 ? 1 : 2);"},
	"syr": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"tzm": {PluralForms: "nplurals=2; plural=(n >= 2 && (n < 11 || n > 99));"},
	"wae": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ak": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"an": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"bn": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ia": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ka": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"mn": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"pa": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"sq": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"cy": {PluralForms: "nplurals=6; plural=(n == 0 ? 0 : n == 1 ? 1 : n == 2 ? 2 : n == 3 ? 3 : n == 6 ? 4 : 5);"},
	"jgo": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sl": {PluralForms: "nplurals=4; plural=(n % 100 == 1 ? 0 : n % 100 == 2 ? 1 : (n % 100 == 3 || n % 100 == 4) ? 2 : 3);"},
	"zu": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"rm": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"scn": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"seh": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sk": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n >= 2 && n <= 4 ? 1 : 2);"},
	"ve": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"am": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"dsb": {PluralForms: "nplurals=4; plural=(n % 100 == 1 ? 0 : n % 100 == 2 ? 1 : (n % 100 == 3 || n % 100 == 4) ? 2 : 3);"},
	"et": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"fi": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"fr": {PluralForms: "nplurals=3; plural=((n == 0 || n == 1) ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"ku": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"mas": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"nn": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"fy": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sc": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ssy": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"vi": {PluralForms: "nplurals=1; plural=(0);"},
	"gl": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"haw": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"he": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"hr": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"id": {PluralForms: "nplurals=1; plural=(0);"},
	"kl": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ko": {PluralForms: "nplurals=1; plural=(0);"},
	"my": {PluralForms: "nplurals=1; plural=(0);"},
	"bo": {PluralForms: "nplurals=1; plural=(0);"},
	"hi": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ln": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ms": {PluralForms: "nplurals=1; plural=(0);"},
	"nd": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"nyn": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"om": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"rof": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ga": {PluralForms: "nplurals=5; plural=(n == 1 ? 0 : n == 2 ? 1 : n >= 3 && n <= 6 ? 2 : n >= 7 && n <= 10 ? 3 : 4);"},
	"af": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"bg": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"chr": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ee": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"eu": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"fil": {PluralForms: "nplurals=2; plural=(n != 1 && n != 2 && n != 3 && (n % 10 == 4 || n % 10 == 6 || n % 10 == 9));"},
	"gsw": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"cs": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n >= 2 && n <= 4 ? 1 : 2);"},
	"kab": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ksh": {PluralForms: "nplurals=3; plural=(n == 0 ? 0 : n == 1 ? 1 : 2);"},
	"mk": {PluralForms: "nplurals=2; plural=(n % 10 != 1 || n % 100 == 11);"},
	"nr": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sdh": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"smn": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"sn": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"kea": {PluralForms: "nplurals=1; plural=(0);"},
	"brx": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"gu": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"hsb": {PluralForms: "nplurals=4; plural=(n % 100 == 1 ? 0 : n % 100 == 2 ? 1 : (n % 100 == 3 || n % 100 == 4) ? 2 : 3);"},
	"ja": {PluralForms: "nplurals=1; plural=(0);"},
	"nb": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"nnh": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"pcm": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ar": {PluralForms: "nplurals=6; plural=(n == 0 ? 0 : n == 1 ? 1 : n == 2 ? 2 : n % 100 >= 3 && n % 100 <= 10 ? 3 : n % 100 >= 11 && n % 100 <= 99 ? 4 : 5);"},
	"ca": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"ceb": {PluralForms: "nplurals=2; plural=(n != 1 && n != 2 && n != 3 && (n % 10 == 4 || n % 10 == 6 || n % 10 == 9));"},
	"dv": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"en": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"hnj": {PluralForms: "nplurals=1; plural=(0);"},
	"kn": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"st": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"bho": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"guw": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"hu": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"hy": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"kaj": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ny": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ta": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ses": {PluralForms: "nplurals=1; plural=(0);"},
	"tig": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"tr": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"uk": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"bm": {PluralForms: "nplurals=1; plural=(0);"},
	"eo": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"smi": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"su": {PluralForms: "nplurals=1; plural=(0);"},
	"teo": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"vun": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"wa": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"yo": {PluralForms: "nplurals=1; plural=(0);"},
	"az": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"io": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ky": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ne": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"so": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"uz": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"vec": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"gd": {PluralForms: "nplurals=4; plural=((n == 1 || n == 11) ? 0 : (n == 2 || n == 12) ? 1 : (n >= 3 && n <= 10 || n >= 13 && n <= 19) ? 2 : 3);"},
	"fa": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"no": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sat": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"sv": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ti": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ts": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"yi": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ce": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"es": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"ha": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ksb": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"lb": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"nl": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ps": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ro": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 0 || n != 1 && n % 100 >= 1 && n % 100 <= 19 ? 1 : 2);"},
	"asa": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"rwk": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sma": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"ss": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"tn": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"to": {PluralForms: "nplurals=1; plural=(0);"},
	"ast": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"csw": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"iu": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"kkj": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"lkt": {PluralForms: "nplurals=1; plural=(0);"},
	"lv": {PluralForms: "nplurals=3; plural=(n % 10 == 0 || n % 100 >= 11 && n % 100 <= 19 ? 0 : n % 10 == 1 && n % 100 != 11 ? 1 : 2);"},
	"nah": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"prg": {PluralForms: "nplurals=3; plural=(n % 10 == 0 || n % 100 >= 11 && n % 100 <= 19 ? 0 : n % 10 == 1 && n % 100 != 11 ? 1 : 2);"},
	"bal": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"blo": {PluralForms: "nplurals=3; plural=(n == 0 ? 0 : n == 1 ? 1 : 2);"},
	"kde": {PluralForms: "nplurals=1; plural=(0);"},
	"kk": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"mgo": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"pt": {PluralForms: "nplurals=3; plural=((n == 0 || n == 1) ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"ru": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"sh": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"be": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"br": {PluralForms: "nplurals=5; plural=(n % 10 == 1 && n % 100 != 11 && n % 100 != 71 && n % 100 != 91 ? 0 : n % 10 == 2 && n % 100 != 12 && n % 100 != 72 && n % 100 != 92 ? 1 : ((n % 10 == 3 || n % 10 == 4) || n % 10 == 9) && (n % 100 < 10 || n % 100 > 19) && (n % 100 < 70 || n % 100 > 79) && (n % 100 < 90 || n % 100 > 99) ? 2 : n != 0 && n % 1000000 == 0 ? 3 : 4);"},
	"is": {PluralForms: "nplurals=2; plural=(n % 10 != 1 || n % 100 == 11);"},
	"os": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"tl": {PluralForms: "nplurals=2; plural=(n != 1 && n != 2 && n != 3 && (n % 10 == 4 || n % 10 == 6 || n % 10 == 9));"},
	"xog": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"lt": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && (n % 100 < 11 || n % 100 > 19) ? 0 : n % 10 >= 2 && n % 10 <= 9 && (n % 100 < 11 || n % 100 > 19) ? 1 : 2);"},
	"ars": {PluralForms: "nplurals=6; plural=(n == 0 ? 0 : n == 1 ? 1 : n == 2 ? 2 : n % 100 >= 3 && n % 100 <= 10 ? 3 : n % 100 >= 11 && n % 100 <= 99 ? 4 : 5);"},
	"da": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"it": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n != 0 && n % 1000000 == 0 ? 1 : 2);"},
	"mr": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"osa": {PluralForms: "nplurals=1; plural=(0);"},
	"te": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"th": {PluralForms: "nplurals=1; plural=(0);"},
	"bem": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ii": {PluralForms: "nplurals=1; plural=(0);"},
	"jmc": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ur": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"wo": {PluralForms: "nplurals=1; plural=(0);"},
	"lij": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"km": {PluralForms: "nplurals=1; plural=(0);"},
	"pap": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sd": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sr": {PluralForms: "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"lg": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"mg": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"ckb": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"el": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ig": {PluralForms: "nplurals=1; plural=(0);"},
	"ks": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"se": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"si": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"jv": {PluralForms: "nplurals=1; plural=(0);"},
	"kcg": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"lo": {PluralForms: "nplurals=1; plural=(0);"},
	"cgg": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"de": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"dz": {PluralForms: "nplurals=1; plural=(0);"},
	"fur": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"or": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"jbo": {PluralForms: "nplurals=1; plural=(0);"},
	"nso": {PluralForms: "nplurals=2; plural=(n > 1);"},
	"pl": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	"saq": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"vo": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"fo": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"ml": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"sah": {PluralForms: "nplurals=1; plural=(0);"},
	"sms": {PluralForms: "nplurals=3; plural=(n == 1 ? 0 : n == 2 ? 1 : 2);"},
	"xh": {PluralForms: "nplurals=2; plural=(n != 1);"},
	"as": {PluralForms: "nplurals=2; plural=(n > 1);"},
}
