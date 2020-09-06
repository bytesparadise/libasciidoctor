package parser

//go:generate pigeon -optimize-parser -alternate-entrypoints AsciidocRawDocument,RawFile,TextDocument,DocumentRawBlock,FileLocation,IncludedFileLine,InlineLinks,LabeledListItemTerm,NormalBlockContentSubstitution,VerseBlockContentSubstitution,MarkdownQuoteBlockAttribution,InlineElements,QuotedTextSubstitution,NoneSubstitution,AttributesSubstitution,ReplacementsSubstitution,PostReplacementsSubstitution,InlinePassthroughSubstitution -o parser.go parser.peg
