import os
import re

def replace_in_file(filepath, replacements_map):
    """
    Replaces text in a file based on the provided dictionary of replacements.
    Uses regex for word-boundary-aware and literal string replacements.
    """
    try:
        with open(filepath, 'r', encoding='utf-8', errors='ignore') as f:
            content = f.read()

        new_content = content
        modified = False

        # To ensure longer, more specific patterns are matched before shorter, general ones
        # (e.g., "Vectora" before "Vectora"), sort keys by length in descending order.
        sorted_keys = sorted(replacements_map.keys(), key=len, reverse=True)

        for old_term_raw in sorted_keys:
            new_term = replacements_map[old_term_raw]

            # Construct regex pattern for replacement.
            # Use re.escape to treat the old_term_raw as a literal string in regex.
            # For terms containing spaces or special characters, direct literal match is better.
            # For single words, use word boundaries (\b) to prevent partial word matches.
            
            # This logic assumes old_term_raw might contain spaces, in which case \b isn't appropriate.
            # It also considers different case sensitivity for the terms as provided in replacements_map.
            # The current replacements_map contains specific case variations, so we'll do case-sensitive
            # replacement for those, but ensure it's a whole word match for single words.

            if re.search(r'\s', old_term_raw): # If the term contains whitespace, treat it as a phrase
                # Match the exact phrase, case-sensitively as provided in the map
                pattern = re.compile(re.escape(old_term_raw))
            else:
                # For single words, use word boundaries for precise replacement
                pattern = re.compile(r'\b' + re.escape(old_term_raw) + r'\b')
            
            if pattern.search(new_content):
                new_content = pattern.sub(new_term, new_content)
                modified = True

        if modified:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print(f"Modified: {filepath}")
            return True
        return False

    except Exception as e:
        print(f"Error processing {filepath}: {e}")
        return False

def main(directory):
    # Define the replacements. Ensure all relevant case variations are covered here.
    replacements = {
        # Longer, more specific terms first
        "Vectora": "Vectora",
        "vectora": "vectora",
        "VECTORA": "VECTORA",
        "Vectora": "Vectora",
        "vectora": "vectora",
        "VECTORA": "VECTORA",
        
        # Shorter, more general terms next (will be processed after longer ones due to sorting)
        "Vectora": "Vectora",
        "vectora": "vectora",
        "VECTORA": "VECTORA",
    }

    processed_count = 0
    # Add or remove file extensions as needed for the project.
    file_extensions = (
        '.go', '.md', '.py', '.json', '.yml', '.yaml', '.txt', '.toml', '.mod', '.sum',
        '.sh', '.ps1', '.ts', '.js', '.jsx', '.tsx', '.css', '.html', '.xml', '.vue', '.svelte',
        '.java', '.cs', '.cpp', '.h', '.c', '.php', '.rb', '.swift', '.kt', '.r', '.scala',
        '.groovy', '.ini', '.cfg', '.conf', '.properties', '.env', '.dockerfile', '.lock',
        '.gradle', '.proj', '.csproj', '.sln', '.editorconfig', '.gitattributes', '.gitignore'
    )
    
    # Exclude specific directories to avoid modifying version control files, build artifacts, etc.
    excluded_dirs = (
        '.git', '.vscode', '.next', 'node_modules', '__pycache__', 'dist', 'build',
        'vendor', 'target', 'bin', '.idea', '.DS_Store', 'logs', 'tmp', 'temp',
        'out', 'coverage', 'Generated' # Common build/temp directories
    )

    print(f"Starting replacement process in directory: {directory}")
    print(f"Targeting files with extensions: {', '.join(file_extensions)}")
    print(f"Excluding directories: {', '.join(excluded_dirs)}")

    for root, dirs, files in os.walk(directory):
        # Modify dirs in-place to skip excluded directories for os.walk
        dirs[:] = [d for d in dirs if d not in excluded_dirs]

        for filename in files:
            # Check if file has one of the target extensions
            if any(filename.lower().endswith(ext) for ext in file_extensions):
                filepath = os.path.join(root, filename)
                if replace_in_file(filepath, replacements):
                    processed_count += 1
    
    print(f"\nFinished replacement. {processed_count} files were modified.")

if __name__ == "__main__":
    # The script will run from the project root directory.
    target_directory = os.getcwd()
    main(target_directory)
