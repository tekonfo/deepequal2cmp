# deepequal2cmp
gotestsパッケージで自動生成されるDeeqEqualをcmp.Diffに自動変換する。

## demo
![output](https://user-images.githubusercontent.com/31615118/92195271-7ea65300-eea7-11ea-9a86-7e0768719eba.gif)

## why
社内プロジェクトで、go-cmpを利用することがルールづけられているが、手動でいちいち変換しなければならないため。
これを利用することで

- ルールを強制させる
- 手間を省く

ことを実現する。

## how to install
```
go get -u github.com/google/go-cmp/cmp
go get github.com/tekonfo/deepequal2cmp/cmd/deepequal2cmp
```

## how to use

```
deepequal2cmp -d dirpath
```

を実行するとdir内の全ての*_test.goファイルが書き換えられる。

引数が空だとカレントディレクトリで実行される

## 網羅性の確認

gotestsのこの箇所を参照して実装に漏れがないかを確認する

https://github.com/cweill/gotests/blob/develop/internal/render/templates/function.tmpl

チェックする項目としては
- 返り値が複数個の場合
- 返り値が一つのみで、基本系でない場合

## 工夫点
