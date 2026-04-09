# Web-01 源码审计：双层伪装

## 题目描述

你进入了一个“看起来很普通”的页面。表面信息几乎没有价值，但源码里藏着关键信息。

请通过查看以下页面源码片段，找出真正的线索并恢复 flag。

```html
<!-- flag: FAKE_FLAG_IGNORE_THIS -->
<!-- flag: FAKE_FLAG_NOT_HERE -->
<!-- hint: read every suspicious variable carefully -->
<script>
var _0xe02d="Tk9UX1RIRV9GTEFHXzAw";
var _0x4e93="Tk9UX1RIRV9GTEFHXzAx";
var _0x9586="S0VFUF9MT09LSU5H";
var _0x00fa="QUxNT1NUX1RIRVJF";
var _0xde02="Tk9UX1RIRV9GTEFHXzAy";
var _0x2035="Tk9UX1RIRV9GTEFHXzAz";
var _0xcdf8="Tk9UX1RIRV9GTEFHXzA0";
var _0x4a3f="Wm14aFozdDNaV0l0YzI5MWNtTmxMV0YxWkdsMExXUnZkV0pzWlMxM2NtRndMVEF4ZlE9PQ==";
// Y2hlY2sgc3VzcGljaW91cyB2YXJpYWJsZXMgY2FyZWZ1bGx5
</script>
```

## 目标

1. 检查 HTML / JS 源码中的可疑内容
2. 识别真实线索并还原编码数据
3. 还原隐藏的 flag 并提交

## 提示

- 以 `_0x` 开头的变量值得重点关注
- 部分内容可能经过不止一层编码
