# architecture

| ディレクトリ        | 説明                                                                                                                                                                                                          |
|:-------------------:|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
| controllers         | [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) から kick される Reconcile メソッド (エントリーポイント) が定義されており、この中から models, repositories または services を叩く |
| domain/models       | ドメインモデル                                                                                                                                                                                                |
| domain/repositories | ドメインオブジェクトを取得・保存するインタフェース                                                                                                                                                            |
| domain/services     | 複数の外部サービスと連携するロジック (cf. DB アクセスとそのキャッシュ管理) が定義されている。controller からのみ呼び出され、この中から repositories または models を叩く                                      |
| gateways/           | 外部サービスを叩く最小機能のメソッドが定義されており、それぞれ外部サービスごとにパッケージが分割されている                                                                                                    |
| .                   | main関数。controller-manager や各種リポジトリの初期化を実施                                                                                                                                                   |
| wire                | https://github.com/google/wire を利用して DI を管理するパッケージ                                                                                                                                             |
| utils               | どのパッケージからも依存されうるユーティリティ関数置き場                                                                                                                                                      |
| errors              | どのパッケージからも依存されうるエラーコード置き場                                                                                                                                                            |

## 参考

* https://christina04.hatenablog.com/entry/go-clean-architecture のパッケージ分けを特に参考にさせてもらいました。

## FAQ

### models が api に依存してしまっていて良いのか？

yes.

domain/models が api/v1alpha1 に依存しているため一見 apiVersion を更新する際に困るようにみえるが、K8s の ConversionWebhook 機能によりコントローラは最新の 1 apiVersion のみサポートすればよいため、models 以下の依存先を api/v1beta1 に変えれば良いと考えている。
とはいえ api にドメインモデルが依存するのは良くない (モデルが安定でない) ので、 apiVersion を更新する作業が発生した際に models を再考慮すると良いかもしれない。

### usecase 層は無いのか？

現状 usecase 層は明確にパッケージを切っておらず、 controllers パッケージ以下の *_phase.go が usecase の役割を担っている。
パッケージを分けていない理由として以下がある。

* 実装の参考としていた [cluster-api v1.0.5](https://github.com/kubernetes-sigs/cluster-api/tree/v1.0.5/controllers) がそのような作りになっていたから
* controller 層と usecase 層が同パッケージにあったほうが変数の取り回し等の点で実装を書くのが楽だから
