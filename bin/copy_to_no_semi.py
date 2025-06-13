import os,sys

def remove_trailing_semicolons(src_path, dst_path):
    with open(src_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    with open(dst_path, 'w', encoding='utf-8') as f:
        for line in lines:
            # Remove trailing semicolon if present (ignoring whitespace)
            stripped = line.rstrip()
            if stripped.endswith(';'):
                # Remove only the last semicolon, preserve indentation and comments
                idx = stripped.rfind(';')
                stripped = stripped[:idx] + stripped[idx+1:]
            f.write(stripped + '\n')

def process_dir(root):
    for dirpath, _, filenames in os.walk(root):
        for filename in filenames:
            if filename.endswith('.lox') and not filename.endswith('_ns.lox'):
                src = os.path.join(dirpath, filename)
                dst = os.path.join(dirpath, filename[:-4] + '_ns.lox')
                remove_trailing_semicolons(src, dst)
                print(f'Processed: {src} -> {dst}')

if __name__ == '__main__':
    # Change '.' to your target directory if needed
    process_dir(sys.argv[1] if len(sys.argv) > 1 else '.')
