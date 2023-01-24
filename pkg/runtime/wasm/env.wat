;; module env should be exported as "env".
;;
;; This should be regenerated via `wat2wasm --debug-names env.wat`
;;
;; The docs and signatures match used here match the following:
;; https://github.com/aduros/wasm4/blob/main/cli/assets/templates/c/src/wasm4.h
;; https://github.com/aduros/wasm4/blob/main/runtimes/native/src/runtime.h
(module $env

;; ┌──────────────────────────────────────────────────────────────────────────┐
;; │                                                                          │
;; │ Drawing Functions                                                        │
;; │                                                                          │
;; └──────────────────────────────────────────────────────────────────────────┘

  ;; blit copies pixels to the framebuffer.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func $blit (export "blit") (import "goenv" "blit")
    (param $sprite i32)
    (param $x i32)
    (param $y i32)
    (param $width i32)
    (param $height i32)
    (param $flags i32))

  ;; blitSub copies a subregion within a larger sprite atlas to the
  ;; framebuffer.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func $blitSub (export "blitSub") (import "goenv" "blitSub")
    (param $sprite i32)
    (param $x i32)
    (param $y i32)
    (param $width i32)
    (param $height i32)
    (param $srcX i32)
    (param $srcY i32)
    (param $stride i32)
    (param $flags i32))

  ;; line draws a line between two points.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "line") (import "goenv" "line")
    (param $x1 i32)
    (param $y1 i32)
    (param $x2 i32)
    (param $y2 i32))

  ;; hline draws a horizontal line.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "hline") (import "goenv" "hline")
    (param $x i32)
    (param $y i32)
    (param $len i32))
    
  ;; vline draws a vertical line.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "vline") (import "goenv" "vline")
    (param $x i32)
    (param $y i32)
    (param $len i32))

  ;; oval draws an oval (or circle).
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "oval") (import "goenv" "oval")
    (param $x i32)
    (param $y i32)
    (param $width i32)
    (param $height i32))

  ;; rect draws a rectangle.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "rect") (import "goenv" "rect")
    (param $x i32)
    (param $y i32)
    (param $width i32)
    (param $height i32))

  ;; text draws text using the built-in system font from a  *zero-terminated*
  ;; string pointer.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "text") (import "goenv" "text")
    (param $str i32)
    (param $x i32)
    (param $y i32))

  ;; textUtf8 draws text using the built-in system font from a UTF-8 encoded
  ;; input.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "textUtf8") (import "goenv" "textUtf8")
    (param $str i32)
    (param $byteLength i32)
    (param $x i32)
    (param $y i32))

  ;; textUtf16 draws text using the built-in system font from a UTF-16 encoded
  ;; input.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "textUtf16") (import "goenv" "textUtf16")
    (param $str i32)
    (param $byteLength i32)
    (param $x i32)
    (param $y i32))


;; ┌──────────────────────────────────────────────────────────────────────────┐
;; │                                                                          │
;; │ Sound Functions                                                          │
;; │                                                                          │
;; └──────────────────────────────────────────────────────────────────────────┘

  ;; tone plays a sound tone.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "tone") (import "goenv" "tone")
    (param $frequency i32)
    (param $duration i32)
    (param $volume i32)
    (param $flags i32))

;; ┌──────────────────────────────────────────────────────────────────────────┐
;; │                                                                          │
;; │ Storage Functions                                                        │
;; │                                                                          │
;; └──────────────────────────────────────────────────────────────────────────┘

  ;; diskr reads up to `size` bytes from persistent storage into the pointer
  ;; `dest`.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "diskr") (import "goenv" "diskr")
    (param $dest i32)
    (param $size i32)
    (result (; $bytesRead ;)i32))

  ;; diskw writes up to `size` bytes from the pointer `src` into persistent
  ;; storage.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "diskw") (import "goenv" "diskw")
    (param $src i32)
    (param $size i32)
    (result (; $bytesWritten ;) i32))

;; ┌──────────────────────────────────────────────────────────────────────────┐
;; │                                                                          │
;; │ Other Functions                                                          │
;; │                                                                          │
;; └──────────────────────────────────────────────────────────────────────────┘

  ;; trace prints a message to the debug console from a *zero-terminated*
  ;; string pointer.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "trace") (import "goenv" "trace")
    (param $str i32))

  ;; traceUtf8 prints a message to the debug console from a UTF-8 encoded
  ;; input.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "traceUtf8") (import "goenv" "traceUtf8")
    (param $str i32)
    (param $byteLength i32))

  ;; traceUtf16 prints a message to the debug console from a UTF-16 encoded
  ;; input.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "traceUtf16") (import "goenv" "traceUtf16")
    (param $str i32)
    (param $byteLength i32))

  ;; tracef prints a message to the debug console from the following input:
  ;;
  ;; * %c, %d, and %x expect 32-bit integers.
  ;; * %f expects 64-bit floats.
  ;; * %s expects a *zero-terminated* string pointer.
  ;;
  ;; Note: This re-exports the same function from the "goenv" module.
  (func (export "tracef") (import "goenv" "tracef")
    (param $str i32)
    (param $stack i32))

  ;; export 64KB (1 page) to the cartridge module
  (memory (export "memory") 1 1)
)