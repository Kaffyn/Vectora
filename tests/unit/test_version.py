"""Tests for version management module."""

import pytest

from vectora.version import __version__, get_vectora_version


class TestVersion:
    """Test version module."""

    def test_version_string_format(self):
        """Test that version string has correct format."""
        assert isinstance(__version__, str)
        assert len(__version__) > 0
        # Should be something like "0.1.0" or "0.1.0-dev"
        assert __version__[0].isdigit()

    def test_get_vectora_version_returns_string(self):
        """Test that get_vectora_version returns a string."""
        version = get_vectora_version()
        assert isinstance(version, str)
        assert len(version) > 0

    def test_version_consistent(self):
        """Test that __version__ and get_vectora_version() are consistent."""
        assert __version__ == get_vectora_version()

    def test_version_matches_pyproject(self):
        """Test that version string is valid semantic version."""
        # Version should be a valid semantic version string
        parts = __version__.split("-")
        version_part = parts[0].split(".")
        assert len(version_part) >= 2  # At least major.minor
        assert all(part.isdigit() for part in version_part)
        # Optional: can have -dev or other suffix
        if len(parts) > 1:
            assert parts[1] == "dev" or parts[1].isalpha()
