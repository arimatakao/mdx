{
  lib,
  buildGoModule,
  go,
}:
let
  inherit (builtins)
    readFile
    head
    filter
    map
    match
    isList
    isNull
    split
    ;
  inherit (lib.lists) flatten;
  metaGo = readFile ../app/meta.go;
  metaGoLines = filter (x: !(isList x)) (split "[[:space:]]+" metaGo);
  version = head (flatten (filter (x: !(isNull x)) (map (x: match ''"v([.0-9]+)"'' x) metaGoLines)));
in
buildGoModule {
  pname = "mdx";
  version = version;

  src = builtins.path {
    name = "mdx-source";
    path = ./..;
  };

  nativeBuildInputs = [ go ];

  vendorHash = "sha256-+q8OQNU43E0Hdpinj+FSZdf+Nirdemc/Is4mZC8YAcs=";

  # test suite requires network
  doCheck = false;

  meta = with lib; {
    description = "mdx is a command-line interface program for downloading manga from the MangaDex website. The program uses MangaDex API to fetch manga content. ";
    homepage = "https://github.com/arimatakao/mdx";
    changelog = "https://github.com/arimatakao/mdx";
    license = licenses.mit;
    maintainers = with maintainers; [ ];
    mainProgram = "mdx";
  };
}
