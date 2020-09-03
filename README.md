# deepequal2cmp
testingパッケージで自動生成されるDeeqequalをgo-cmpに自動変換する。

## why
社内プロジェクトで、go-cmpを利用することがルールづけられているが、手動でいちいち変換しなければならないため。
これを利用することで

- ルールを強制させる
- 手間を省く

ことを実現する。

## how to use
```
de2cmp ./...
```

を実行すると全ての*_test.goファイルが書き換えられる。

※ 現状は/cmd/deepequal2cmp/main.goをrunしてデバックしています

## 網羅性の確認

gotestsのこの箇所を参照して実装に漏れがないかを確認する

https://github.com/cweill/gotests/blob/develop/internal/render/templates/function.tmpl

チェックする項目としては
- 返り値が複数個の場合
- 返り値が一つのみで、基本系でない場合
