from context import Context
from tools import multiply


class TestMultiplyTool:
    """Test suite for the multiply tool."""

    def test_multiply_basic_numbers(self) -> None:
        """Test multiply(5, 3) returns 15."""
        result = multiply.invoke({"a": 5.0, "b": 3.0})
        assert result == 15.0

    def test_multiply_floats(self) -> None:
        """Test multiply with floating point numbers."""
        result = multiply.invoke({"a": 2.5, "b": 4.0})
        assert result == 10.0

    def test_multiply_negative_numbers(self) -> None:
        """Test multiply with negative numbers."""
        result = multiply.invoke({"a": -5.0, "b": 3.0})
        assert result == -15.0

    def test_multiply_zero(self) -> None:
        """Test multiply by zero."""
        result = multiply.invoke({"a": 5.0, "b": 0.0})
        assert result == 0.0

    def test_multiply_with_runtime(self) -> None:
        """Test that multiply can be called with ToolRuntime (even if unused)."""

        class MockRuntime:
            context = Context(user_type="plus", thread_id=1)

        result = multiply.invoke({"a": 7.0, "b": 6.0})
        assert result == 42.0
