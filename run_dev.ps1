# 개발 모드(Windows) — run_dev.sh 의 윈도우 판. run_dev.bat 을 더블클릭하면 이 파일이 돈다.
#   필요한 도구 확인(없으면 설치 — Go·MariaDB 는 관리자 권한 없이 %LOCALAPPDATA%\zbd-tools 에) → 설정·DB 준비 → BE(Go)·FE(Svelte) 를 띄운다.
#   BE 는 따로 열리는 창([be])에서, FE 는 이 창에서 돈다. 이 창에서 Ctrl+C 로 끄면 BE 창도 같이 꺼진다. DB(MariaDB)는 뒤에서 계속 돈다.
#
# 처음 한 번 만드는 것(모두 git 에 올리지 않는다): zbd-be-golang\.env.local(키 무작위 — 잃으면 회원 개인정보를 못 읽는다) · .dev\account.txt(첫 관리자·시험 회원) · .dev\mariadb(DB 파일)
# 포트: BE 5001 · FE 5000 · DB 3306 (환경변수 BE_PORT·FE_PORT·DB_PORT 로 바꿀 수 있다)
$ErrorActionPreference = 'Stop'
try { [Console]::OutputEncoding = [Text.Encoding]::UTF8 } catch {}
$OutputEncoding = New-Object System.Text.UTF8Encoding $false

$ROOT = $PSScriptRoot
$BE = Join-Path $ROOT 'zbd-be-golang'
$FE = Join-Path $ROOT 'zbd-fe-svelte5'
$DEV = Join-Path $ROOT '.dev'
$TOOLS = Join-Path $env:LOCALAPPDATA 'zbd-tools'
$BE_PORT = if ($env:BE_PORT) { $env:BE_PORT } else { '5001' }
$FE_PORT = if ($env:FE_PORT) { $env:FE_PORT } else { '5000' }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { '3306' }
$NODE_MIN = 22

function Say($m) { Write-Host "▶ $m" -ForegroundColor Cyan }
function Die($m) { Write-Host "✗ $m" -ForegroundColor Red; exit 1 }
function Fwd($p) { $p -replace '\\', '/' }  # .env 의 경로는 / 로(Go·Node 모두 읽는다)
function Rnd($n) {
  $chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'.ToCharArray()
  $b = New-Object byte[] $n; [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b)
  -join ($b | ForEach-Object { $chars[$_ % $chars.Length] })
}
function B64 { $b = New-Object byte[] 32; [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b); [Convert]::ToBase64String($b) }
function Have($c) { [bool](Get-Command $c -ErrorAction SilentlyContinue) }
# 이 창의 PATH 앞에 넣고, 사용자 PATH 에도 남긴다(다음에 연 창에서도 보이게)
function Add-Path($dir) {
  if (-not (($env:Path -split ';') -contains $dir)) { $env:Path = "$dir;$env:Path" }
  $up = [Environment]::GetEnvironmentVariable('Path', 'User')
  if (-not (($up -split ';') -contains $dir)) { [Environment]::SetEnvironmentVariable('Path', $(if ($up) { "$up;$dir" } else { $dir }), 'User') }
}
function Get-File($url, $out) {
  curl.exe -fsSL -o $out $url
  if ($LASTEXITCODE -ne 0) { Die "받지 못했습니다: $url" }
}
function Port-Busy($p) {
  $c = New-Object Net.Sockets.TcpClient
  try { $c.Connect('127.0.0.1', [int]$p); $true } catch { $false } finally { $c.Close() }
}

# ─── 1. 도구 ───────────────────────────────────────────────
Say '도구 확인'
New-Item -ItemType Directory -Force -Path $TOOLS | Out-Null
if (Test-Path "$TOOLS\go\bin\go.exe") { Add-Path "$TOOLS\go\bin" }
if (-not (Have go)) {
  Say 'Go 설치(관리자 권한 없이 — %LOCALAPPDATA%\zbd-tools\go)'
  $rel = (Invoke-RestMethod 'https://go.dev/dl/?mode=json')[0]
  $f = $rel.files | Where-Object { $_.os -eq 'windows' -and $_.arch -eq 'amd64' -and $_.kind -eq 'archive' } | Select-Object -First 1
  $zip = Join-Path $env:TEMP $f.filename
  Get-File "https://go.dev/dl/$($f.filename)" $zip
  if ((Get-FileHash $zip -Algorithm SHA256).Hash.ToLower() -ne $f.sha256) { Die 'Go 파일 체크섬이 맞지 않습니다' }
  Expand-Archive -Force $zip $TOOLS
  Remove-Item $zip
  Add-Path "$TOOLS\go\bin"
}
$GOBIN = Join-Path (go env GOPATH) 'bin'
Add-Path $GOBIN

$nodeOk = $false
if (Have node) { $nodeOk = [int](node -p "process.versions.node.split('.')[0]") -ge $NODE_MIN }
if (-not $nodeOk) {
  Say "Node.js $NODE_MIN 이상 설치(winget — 설치 창이 뜨면 허용)"
  winget install -e --id OpenJS.NodeJS.LTS --accept-source-agreements --accept-package-agreements
  $env:Path = [Environment]::GetEnvironmentVariable('Path', 'Machine') + ';' + [Environment]::GetEnvironmentVariable('Path', 'User')
  if (-not (Have node)) { Die 'Node.js 를 설치한 뒤 이 창을 닫고 run_dev.bat 을 다시 실행하세요' }
}

$MDB = Join-Path $TOOLS 'mariadb'
if (-not (Test-Path "$MDB\bin\mariadbd.exe")) {
  Say 'MariaDB 설치(관리자 권한 없이 — %LOCALAPPDATA%\zbd-tools\mariadb)'
  $rel = Invoke-RestMethod 'https://downloads.mariadb.org/rest-api/mariadb/11.8/latest'
  $file = $null
  foreach ($r in $rel.releases.PSObject.Properties.Value) { foreach ($x in $r.files) { if ($x.os -eq 'Windows' -and $x.file_name -like '*-winx64.zip') { $file = $x } } }
  if (-not $file) { Die 'MariaDB 내려받기 주소를 찾지 못했습니다' }
  $zip = Join-Path $env:TEMP $file.file_name
  Get-File ($file.file_download_url -replace '^http:', 'https:') $zip
  if ($file.checksum.sha256sum -and (Get-FileHash $zip -Algorithm SHA256).Hash.ToLower() -ne $file.checksum.sha256sum.ToLower()) { Die 'MariaDB 파일 체크섬이 맞지 않습니다' }
  $tmp = Join-Path $env:TEMP 'zbd-mariadb'
  Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $tmp
  Expand-Archive -Force $zip $tmp
  Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $MDB
  Move-Item (Get-ChildItem $tmp -Directory | Select-Object -First 1).FullName $MDB
  Remove-Item -Recurse -Force $tmp, $zip
}
Add-Path "$MDB\bin"

if (-not (Test-Path "$GOBIN\air.exe")) { Say 'air 설치(Go 저장 시 다시 시작)'; go install github.com/air-verse/air@latest }
Write-Host "  $(go env GOVERSION) · node $(node -v) · $((& "$MDB\bin\mariadb.exe" --version) -replace '.*from ([0-9.]+).*', 'mariadb $1')"

# ─── 2. MariaDB 켜기 ──────────────────────────────────────
$DATA = Join-Path $DEV 'mariadb'
# 네이티브 명령의 오류 출력을 버릴 때(2>$null)는 잠깐 Continue — Stop 이면 PowerShell 5.1 이 그 출력을 오류로 보고 스크립트를 멈춘다
function Quiet([scriptblock]$b) { $old = $ErrorActionPreference; $ErrorActionPreference = 'Continue'; try { & $b 2>$null | Out-Null } finally { $ErrorActionPreference = $old } }
function Db-Up { Quiet { & "$MDB\bin\mariadb-admin.exe" --skip-ssl -uroot --host=127.0.0.1 -P $DB_PORT ping }; $LASTEXITCODE -eq 0 }
if (-not (Test-Path "$DATA\my.ini")) {
  Say 'MariaDB 데이터 폴더 만들기(.dev\mariadb)'
  New-Item -ItemType Directory -Force -Path $DEV | Out-Null
  & "$MDB\bin\mariadb-install-db.exe" "--datadir=$DATA" "--port=$DB_PORT" | Out-Null
  if ($LASTEXITCODE -ne 0) { Die 'MariaDB 데이터 폴더를 만들지 못했습니다' }
  # 이 컴퓨터 안에서만 받는다(방화벽 창이 뜨지 않게, 밖에서 못 붙게)
  $ini = Get-Content "$DATA\my.ini" -Raw
  if ($ini -notmatch 'bind-address') { $ini = $ini -replace '\[mysqld\]', "[mysqld]`r`nbind-address=127.0.0.1"; Set-Content -Encoding ASCII "$DATA\my.ini" $ini }
}
if (-not (Db-Up)) {
  if (Port-Busy $DB_PORT) { Die "포트 $DB_PORT 를 다른 프로그램(다른 MySQL·MariaDB?)이 쓰고 있습니다 — 끄거나 DB_PORT 를 바꾸세요" }
  Say 'MariaDB 시작(창 없이 뒤에서)'
  Start-Process -WindowStyle Hidden -FilePath "$MDB\bin\mariadbd.exe" -ArgumentList "--defaults-file=`"$DATA\my.ini`""
  for ($i = 0; $i -lt 30 -and -not (Db-Up); $i++) { Start-Sleep 1 }
  if (-not (Db-Up)) { Die "MariaDB 가 뜨지 않습니다 — $DATA 의 *.err 파일을 확인하세요" }
}

# ─── 3. 설정 (처음 한 번) ─────────────────────────────────
New-Item -ItemType Directory -Force -Path "$DEV\upload" | Out-Null
$ENVF = Join-Path $BE '.env.local'
if (-not (Test-Path $ENVF)) {
  Say 'BE 설정 만들기(.env.local — 키는 무작위)'
  $lines = @(
    '# run_dev.ps1 이 만든 개발 설정 — git 에 올리지 않는다. 키를 잃으면 이 DB 의 회원 개인정보를 못 읽는다.',
    "PORT=$BE_PORT", 'DB_HOST=127.0.0.1', "DB_PORT=$DB_PORT", 'DB_NAME=zbd_starter', 'DB_USER=zbd_dev', "DB_PASSWORD=$(Rnd 24)",
    'DB_CHARSET=utf8mb4', 'DB_CONN_LIMIT=10', "AES_256_ENCRYPTION_KEY=$(Rnd 32)", "AES_256_ENCRYPTION_IV=$(Rnd 16)",
    "PII_ENC_KEY_V1=$(B64)", "PII_INDEX_KEY=$(B64)", 'AUTH_COOKIE_KEY=zbd_admin_session', 'AUTH_COOKIE_KEY_USER=zbd_user_session',
    "CDN_UPLOAD_DIR=$(Fwd "$DEV\upload")/", "FE_INDEX_HTML=$(Fwd "$DEV\index.html")", 'RUN_JOBS=false'
  )
  [IO.File]::WriteAllText($ENVF, ($lines -join "`n") + "`n", (New-Object System.Text.UTF8Encoding $false))
}
if (-not (Test-Path "$DEV\index.html")) {
  [IO.File]::WriteAllText("$DEV\index.html", "<!doctype html>`n<html lang=`"ko`">`n<head><meta charset=`"utf-8`" /></head>`n<body><div id=`"root`"></div></body>`n</html>`n", (New-Object System.Text.UTF8Encoding $false))
}
$envv = @{}
foreach ($l in Get-Content $ENVF -Encoding UTF8) { if ($l -match '^([A-Z0-9_]+)=(.*)$') { $envv[$Matches[1]] = $Matches[2].Trim('"') } }
# 화면은 기본으로 같은 네트워크의 다른 기기에서도 열린다. 이 컴퓨터에서만 = .env.local 에 FE_HOST=localhost. 서버(BE)·DB 는 그대로 이 컴퓨터 안에서만 받는다
if ($envv.FE_HOST -and -not $env:FE_HOST) { $env:FE_HOST = $envv.FE_HOST }

# ─── 4. DB (db/setup.sh 와 같은 일) ───────────────────────
Say 'DB 준비'
$DBN = $envv.DB_NAME; $DBU = $envv.DB_USER; $DBP = $envv.DB_PASSWORD
function Admin-Sql($sql) { & "$MDB\bin\mariadb.exe" --skip-ssl -uroot --host=127.0.0.1 -P $DB_PORT --default-character-set=utf8mb4 -e $sql; if ($LASTEXITCODE -ne 0) { Die 'DB 관리자 명령 실패' } }
function App-Sql([string[]]$a) { $env:MYSQL_PWD = $DBP; try { & "$MDB\bin\mariadb.exe" --skip-ssl --host=127.0.0.1 -P $DB_PORT -u $DBU --default-character-set=utf8mb4 $DBN @a } finally { Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue } }
$env:MYSQL_PWD = $DBP
Quiet { & "$MDB\bin\mariadb.exe" --skip-ssl --host=127.0.0.1 -P $DB_PORT -u $DBU -e "USE ``$DBN``" }
$appOk = $LASTEXITCODE -eq 0
Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue
if (-not $appOk) {
  Write-Host "· DB $DBN · 계정 $DBU 만들기"
  Admin-Sql "CREATE DATABASE IF NOT EXISTS ``$DBN`` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; CREATE USER IF NOT EXISTS '$DBU'@'localhost' IDENTIFIED BY '$DBP'; CREATE USER IF NOT EXISTS '$DBU'@'127.0.0.1' IDENTIFIED BY '$DBP'; ALTER USER '$DBU'@'localhost' IDENTIFIED BY '$DBP'; ALTER USER '$DBU'@'127.0.0.1' IDENTIFIED BY '$DBP'; GRANT ALL PRIVILEGES ON ``$DBN``.* TO '$DBU'@'localhost'; GRANT ALL PRIVILEGES ON ``$DBN``.* TO '$DBU'@'127.0.0.1';"
}
App-Sql @('-e', 'CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(100) NOT NULL PRIMARY KEY, appliedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)')
# 플러그인 빼기(zbd remove)가 남긴 DB 지우기 — 마이그레이션보다 먼저, 파일마다 한 번(db/setup.sh 와 같은 규칙)
foreach ($f in Get-ChildItem "$ROOT\db\removals\*.sql" -ErrorAction SilentlyContinue | Sort-Object Name) {
  $v = 'remove:' + ($f.Name -replace '\.sql$', '')
  $n = App-Sql @('-N', '-e', "SELECT COUNT(*) FROM schema_migrations WHERE version='$v'")
  if ("$n".Trim() -ne '0') { continue }
  $vs = (Select-String -Path $f.FullName -Pattern '^-- migrations: (.*)$' | Select-Object -First 1)
  $had = $true
  if ($vs -and $vs.Matches[0].Groups[1].Value.Trim()) {
    $had = $false
    foreach ($m in $vs.Matches[0].Groups[1].Value.Trim() -split '\s+') {
      $c = App-Sql @('-N', '-e', "SELECT COUNT(*) FROM schema_migrations WHERE version='$m'")
      if ("$c".Trim() -ne '0') { $had = $true }
    }
  }
  if ($had) {
    Write-Host "· 플러그인 빼기 $($f.Name)"
    App-Sql @('-e', "source $(Fwd $f.FullName)")
    if ($LASTEXITCODE -ne 0) { Die "플러그인 빼기 $($f.Name) 실패" }
  }
  App-Sql @('-e', "INSERT INTO schema_migrations (version) VALUES ('$v')")
}
foreach ($f in Get-ChildItem "$ROOT\db\migrations\*.up.sql" | Sort-Object Name) {
  $v = $f.Name -replace '\.up\.sql$', ''
  $n = App-Sql @('-N', '-e', "SELECT COUNT(*) FROM schema_migrations WHERE version='$v'")
  if ("$n".Trim() -ne '0') { continue }
  Write-Host "· 마이그레이션 $v"
  App-Sql @('-e', "source $(Fwd $f.FullName)")
  if ($LASTEXITCODE -ne 0) { Die "마이그레이션 $v 실패" }
  App-Sql @('-e', "INSERT INTO schema_migrations (version) VALUES ('$v')")
}
foreach ($f in Get-ChildItem "$ROOT\db\seed\*.sql" | Sort-Object Name) {
  App-Sql @('-e', "source $(Fwd $f.FullName)")
  if ($LASTEXITCODE -ne 0) { Die "기본 데이터 $($f.Name) 실패" }
}
Write-Host "✓ DB 준비됨 ($DBN)"

# 첫 관리자 · 시험 회원 — 처음 한 번(.dev\account.txt 가 없을 때). 비밀번호는 무작위, 이 파일에만
$ACC = Join-Path $DEV 'account.txt'
if (-not (Test-Path $ACC)) {
  $ap = Rnd 16; $mp = Rnd 16
  Push-Location $BE
  try {
    $env:ADMIN_PASSWORD = $ap; go run ./cmd/create-admin -email admin@example.com -name '관리자' -image (Fwd "$ROOT\db\assets\profile\400x400_profile_admin.webp"); if ($LASTEXITCODE -ne 0) { Die '첫 관리자를 만들지 못했습니다' }
    $env:MEMBER_PASSWORD = $mp; go run ./cmd/create-member -email user@example.com -name '시험회원' -nickname '테스터' -image (Fwd "$ROOT\db\assets\profile\400x400_profile_user.webp"); if ($LASTEXITCODE -ne 0) { Die '시험 회원을 만들지 못했습니다' }
  } finally { Remove-Item Env:ADMIN_PASSWORD, Env:MEMBER_PASSWORD -ErrorAction SilentlyContinue; Pop-Location }
  $txt = "# 개발용 계정 — git 에 올라가지 않는다`n[관리자] http://localhost:$FE_PORT/admin`nemail=admin@example.com`npassword=$ap`n`n[사용자] http://localhost:$FE_PORT/login`nemail=user@example.com`npassword=$mp`n"
  [IO.File]::WriteAllText($ACC, $txt, (New-Object System.Text.UTF8Encoding $false))
}

# 사이트 기본 그림 · 공지 예시 첨부 · 웹하드 시작 가이드(없을 때만)
New-Item -ItemType Directory -Force -Path "$DEV\upload\ZBD\site" | Out-Null
foreach ($f in Get-ChildItem "$ROOT\db\assets\site" -File) { if (-not (Test-Path "$DEV\upload\ZBD\site\$($f.Name)")) { Copy-Item $f.FullName "$DEV\upload\ZBD\site\" } }
$NF = "$DEV\secure\ZBD\post_notice01"
if (-not (Test-Path "$NF\시작 가이드.md")) { New-Item -ItemType Directory -Force -Path $NF | Out-Null; Copy-Item "$ROOT\docs\GETTING_STARTED.md" "$NF\시작 가이드.md" }
$WH = "$DEV\file-explorer\ZBD\WEB_HDD"
if (-not (Test-Path $WH)) { New-Item -ItemType Directory -Force -Path $WH | Out-Null; Copy-Item "$ROOT\docs\GETTING_STARTED.md" "$WH\시작 가이드.md"; Say '웹하드에 시작 가이드 넣음' }

# ─── 5. FE 패키지 ────────────────────────────────────────
$lock = Join-Path $FE 'package-lock.json'; $mark = Join-Path $FE 'node_modules\.package-lock.json'
if (-not (Test-Path $mark) -or (Get-Item $lock).LastWriteTime -gt (Get-Item $mark).LastWriteTime) {
  Say 'FE 패키지 설치(npm ci)'
  Push-Location $FE; try { npm ci --no-audit --no-fund; if ($LASTEXITCODE -ne 0) { Die 'npm ci 실패' } } finally { Pop-Location }
}

# ─── 5½. 소켓 데몬(웹소켓 플러그인을 붙였을 때만 — plugins/zbd-socket) ─────
$SOCK = Join-Path $ROOT 'plugins\zbd-socket'
$sockProc = $null
if (Test-Path "$SOCK\main.go") {
  foreach ($kv in @(@('SOCKET_URL', 'http://127.0.0.1:4006'), @('SOCKET_SECRET', (Rnd 32)))) {
    if (-not $envv[$kv[0]]) { Add-Content -Path $ENVF -Value "$($kv[0])=$($kv[1])" -Encoding UTF8; $envv[$kv[0]] = $kv[1] }
  }
  $SD = Join-Path $DEV 'socket'
  New-Item -ItemType Directory -Force -Path $SD | Out-Null
  Say '소켓 데몬 빌드(plugins/zbd-socket)'
  Push-Location $SOCK; try { go build -buildvcs=false -o "$SD\zbd-socket.exe" .; if ($LASTEXITCODE -ne 0) { Die '소켓 데몬 빌드 실패' } } finally { Pop-Location }
  # 데몬 설정 — BE .env.local 에서 같은 값을 넘긴다(매번 다시 쓴다). 출처는 이 컴퓨터 + 화면 포트(같은 네트워크의 다른 기기에서 열어도)
  $sockEnv = @("ADMIN_SOCKET_SECRET=$($envv.SOCKET_SECRET)", "AES_256_ENCRYPTION_KEY=$($envv.AES_256_ENCRYPTION_KEY)", "AES_256_ENCRYPTION_IV=$($envv.AES_256_ENCRYPTION_IV)",
    "CLAUDE_BE_URL=http://127.0.0.1:$BE_PORT", "ORIGINS=*:$FE_PORT", "READY_SOCKET_SECRET=$($envv.READY_SOCKET_SECRET)", "READY_QUEUE_HOSTING_IDS=ZBD")
  [IO.File]::WriteAllText("$SD\.env", ($sockEnv -join "`n") + "`n", (New-Object System.Text.UTF8Encoding $false))
  if (Port-Busy 4006) { Write-Host '  포트 4006 에 이미 소켓 데몬이 떠 있습니다 — 그대로 씁니다' }
  else { $sockProc = Start-Process -WindowStyle Hidden -FilePath "$SD\zbd-socket.exe" -WorkingDirectory $SD -PassThru; Write-Host "  소켓 데몬 127.0.0.1:4006 (pid $($sockProc.Id))" }
}

# ─── 6. 실행 ─────────────────────────────────────────────
function Stop-Tree($procId) {
  Get-CimInstance Win32_Process -Filter "ParentProcessId=$procId" -ErrorAction SilentlyContinue | ForEach-Object { Stop-Tree $_.ProcessId }
  Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
}
# Stop-Mine 포트 — 그 포트를 쥔 프로세스 중 이 스타터 폴더에서 뜬 것(명령줄·실행 파일 경로에 폴더가 든 것 — vite·Go 서버)만 끈다. 남의 프로그램은 그대로.
# 창을 ✕ 로 닫거나 강제로 끝내면 vite·서버가 남아 다음 실행이 「포트 사용 중」 으로 막혔다(2026-10-06 리눅스에서 먼저 겪음)
function Stop-Mine($p, $why) {
  $mine = ($ROOT -replace '/', '\').TrimEnd('\').ToLower() + '\'
  $ids = Get-NetTCPConnection -LocalPort $p -State Listen -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
  foreach ($id in $ids) {
    if (-not $id) { continue }
    $w = Get-CimInstance Win32_Process -Filter "ProcessId=$id" -ErrorAction SilentlyContinue
    $cl = ("$($w.CommandLine) $($w.ExecutablePath)" -replace '/', '\').ToLower()
    if ($cl.Contains($mine)) {
      if ($why) { Write-Host "  포트 ${p}: $why(pid $id) — 끕니다" }
      Stop-Tree $id
    }
  }
  for ($i = 0; $i -lt 10 -and (Port-Busy $p); $i++) { Start-Sleep -Milliseconds 500 }
}
foreach ($p in $BE_PORT, $FE_PORT) {
  if (Port-Busy $p) { Stop-Mine $p '지난번 실행이 남긴 프로그램' }
  if (Port-Busy $p) { Die "포트 $p 를 이미 다른 프로그램이 쓰고 있습니다 — 끄거나 BE_PORT/FE_PORT 를 바꾸세요" }
}
Say "BE http://127.0.0.1:$BE_PORT · FE http://localhost:$FE_PORT (관리자 http://localhost:$FE_PORT/admin — 계정은 .dev\account.txt)"
if ($env:FE_HOST -ne 'localhost') {
  $ips = Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue | Where-Object { $_.IPAddress -notlike '127.*' -and $_.IPAddress -notlike '169.254.*' -and $_.PrefixOrigin -ne 'WellKnown' } | Select-Object -ExpandProperty IPAddress
  foreach ($ip in $ips) { Write-Host "  같은 네트워크의 다른 기기: http://${ip}:$FE_PORT" -ForegroundColor Cyan }
  Write-Host '  (방화벽 창이 뜨면 「개인 네트워크」 허용 — 안 열리면 Windows 보안 › 방화벽에서 node.exe 를 허용)'
}
$beCmd = "`$Host.UI.RawUI.WindowTitle='[be] zbd-starter'; [Console]::OutputEncoding=[Text.Encoding]::UTF8; `$env:PORT='$BE_PORT'; Set-Location '$BE'; & '$GOBIN\air.exe' -c .air.windows.toml"
$beProc = Start-Process powershell -ArgumentList '-NoProfile', '-NoExit', '-ExecutionPolicy', 'Bypass', '-Command', $beCmd -PassThru
try {
  $env:ZBD_API = "http://127.0.0.1:$BE_PORT"; $env:FE_PORT = $FE_PORT
  Push-Location $FE
  npx vite dev
} finally {
  Pop-Location
  Write-Host ''; Say '끄는 중…'
  Stop-Tree $beProc.Id
  # 그래도 남은 이 폴더의 vite·Go 서버(air 가 띄운 서버는 창 밖에 남을 수 있다) — 다음 실행이 막히지 않게
  Stop-Mine $FE_PORT; Stop-Mine $BE_PORT
  if ($sockProc) { Stop-Process -Id $sockProc.Id -Force -ErrorAction SilentlyContinue }
  Write-Host '  꺼졌습니다(DB 는 뒤에서 계속 돕니다).'
}
