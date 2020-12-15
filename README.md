# fotolife

## Description

はてなフォトライフのCLIクライアントです。
現在は写真の一括ダウンロードだけができます。

## Install

```shell
go get github.com/pen/fotolife/cmd/fotolife
rehash
```

## Usage

```shell
fotolife dump <hatena-id> [<folder>...]
```

`-t`オプションでトップページの写真もダウンロードします。

## ログイン

`-p <password>` を指定するとログインした状態で動作します。非公開フォルダがある場合に使います。

パスワード以外の認証方法は未対応です。
