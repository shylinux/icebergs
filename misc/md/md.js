Volcanos(chat.ONSYNTAX, {
	md: {
		prefix: {"//": code.COMMENT},
		regexp: {"[A-Z_0-9]+": code.CONSTANT},
		keyword: {
			"package": code.KEYWORD,
			"import": code.KEYWORD,
			"public": code.KEYWORD,
			"private": code.KEYWORD,
			"static": code.KEYWORD,
		},
	},
})