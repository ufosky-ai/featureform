"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[4657],{96412:function(a){function b(a){!function(a){var b=a.util.clone(a.languages.javascript),c=/(?:\s|\/\/.*(?!.)|\/\*(?:[^*]|\*(?!\/))\*\/)/.source,d=/(?:\{(?:\{(?:\{[^{}]*\}|[^{}])*\}|[^{}])*\})/.source,e=/(?:\{<S>*\.{3}(?:[^{}]|<BRACES>)*\})/.source;function f(a,b){return RegExp(a=a.replace(/<S>/g,function(){return c}).replace(/<BRACES>/g,function(){return d}).replace(/<SPREAD>/g,function(){return e}),b)}e=f(e).source,a.languages.jsx=a.languages.extend("markup",b),a.languages.jsx.tag.pattern=f(/<\/?(?:[\w.:-]+(?:<S>+(?:[\w.:$-]+(?:=(?:"(?:\\[\s\S]|[^\\"])*"|'(?:\\[\s\S]|[^\\'])*'|[^\s{'"/>=]+|<BRACES>))?|<SPREAD>))*<S>*\/?)?>/.source),a.languages.jsx.tag.inside.tag.pattern=/^<\/?[^\s>\/]*/,a.languages.jsx.tag.inside["attr-value"].pattern=/=(?!\{)(?:"(?:\\[\s\S]|[^\\"])*"|'(?:\\[\s\S]|[^\\'])*'|[^\s'">]+)/,a.languages.jsx.tag.inside.tag.inside["class-name"]=/^[A-Z]\w*(?:\.[A-Z]\w*)*$/,a.languages.jsx.tag.inside.comment=b.comment,a.languages.insertBefore("inside","attr-name",{spread:{pattern:f(/<SPREAD>/.source),inside:a.languages.jsx}},a.languages.jsx.tag),a.languages.insertBefore("inside","special-attr",{script:{pattern:f(/=<BRACES>/.source),alias:"language-javascript",inside:{"script-punctuation":{pattern:/^=(?=\{)/,alias:"punctuation"},rest:a.languages.jsx}}},a.languages.jsx.tag);var g=function(a){return a?"string"==typeof a?a:"string"==typeof a.content?a.content:a.content.map(g).join(""):""},h=function(b){for(var c=[],d=0;d<b.length;d++){var e=b[d],f=!1;if("string"!=typeof e&&("tag"===e.type&&e.content[0]&&"tag"===e.content[0].type?"</"===e.content[0].content[0].content?c.length>0&&c[c.length-1].tagName===g(e.content[0].content[1])&&c.pop():"/>"===e.content[e.content.length-1].content||c.push({tagName:g(e.content[0].content[1]),openedBraces:0}):c.length>0&&"punctuation"===e.type&&"{"===e.content?c[c.length-1].openedBraces++:c.length>0&&c[c.length-1].openedBraces>0&&"punctuation"===e.type&&"}"===e.content?c[c.length-1].openedBraces--:f=!0),(f||"string"==typeof e)&&c.length>0&&0===c[c.length-1].openedBraces){var i=g(e);d<b.length-1&&("string"==typeof b[d+1]||"plain-text"===b[d+1].type)&&(i+=g(b[d+1]),b.splice(d+1,1)),d>0&&("string"==typeof b[d-1]||"plain-text"===b[d-1].type)&&(i=g(b[d-1])+i,b.splice(d-1,1),d--),b[d]=new a.Token("plain-text",i,null,i)}e.content&&"string"!=typeof e.content&&h(e.content)}};a.hooks.add("after-tokenize",function(a){("jsx"===a.language||"tsx"===a.language)&&h(a.tokens)})}(a)}a.exports=b,b.displayName="jsx",b.aliases=[]}}])