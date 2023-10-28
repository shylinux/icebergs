Volcanos(chat.ONSYNTAX, {
	java: {
		prefix: {"//": code.COMMENT},
		regexp: {"[A-Z_0-9]+": code.CONSTANT},
		keyword: {
			"package": code.KEYWORD,
			"import": code.KEYWORD,
			"public": code.KEYWORD,
			"private": code.KEYWORD,
			"static": code.KEYWORD,
			"final": code.KEYWORD,
			"class": code.KEYWORD,
			"extends": code.KEYWORD,
			"implements": code.KEYWORD,
			"default": code.KEYWORD,

			"new": code.KEYWORD,
			"if": code.KEYWORD,
			"else": code.KEYWORD,
			"for": code.KEYWORD,
			"break": code.KEYWORD,
			"try": code.KEYWORD,
			"catch": code.KEYWORD,
			"return": code.KEYWORD,

			"null": code.CONSTANT,
			"true": code.CONSTANT,
			"false": code.CONSTANT,

			"int": code.DATATYPE,
			"void": code.DATATYPE,
			"string": code.DATATYPE,
			"boolean": code.DATATYPE,
			"Object": code.DATATYPE,
			"Class": code.DATATYPE,
			"String": code.DATATYPE,
			"Integer": code.DATATYPE,
			"Long": code.DATATYPE,
			"Exception": code.DATATYPE,
			"List": code.DATATYPE,
			"Map": code.DATATYPE,

			"this": code.OBJECT,

			"interface": code.FUNCTION,
			"Override": code.FUNCTION,
			"Autowired": code.FUNCTION,
			"Retention": code.FUNCTION,
			"Configuration": code.FUNCTION,
			"Builder": code.FUNCTION,
			"Value": code.FUNCTION,
			"Data": code.FUNCTION,
			"Bean": code.FUNCTION,
			"Service": code.FUNCTION,
			"Controller": code.FUNCTION,
			"Validated": code.FUNCTION,
			"RequestMapping": code.FUNCTION,
			"RequestParam": code.FUNCTION,
			"RequestPart": code.FUNCTION,
			"RequestBody": code.FUNCTION,
			"ResponseBody": code.FUNCTION,
			"PathVariable": code.FUNCTION,
			"ApiOperation": code.FUNCTION,
			"Api": code.FUNCTION,
			"Tag": code.FUNCTION,
		},
		func: function(can, push, text, indent) {
			var ls = can.core.Split(text)
			if (ls[0] == "public") {
				if (ls[1] == "class") {
					push(ls[2])
					return
				}
				for (var i = 0; i < ls.length; i++) {
					if (ls[i] == "(") {
						push("  "+ls[i-1])
						return
					}
				}
			}
		},
	},
})
