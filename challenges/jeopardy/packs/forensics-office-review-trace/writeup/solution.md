# 解法

`.docx` 本质就是 zip。解开后重点看 `word/comments.xml`、`word/document.xml`、`word/footnotes.xml` 等结构化 XML，本题的 flag 直接留在 comments 里。
