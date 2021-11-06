## Blackprint Engine for Golang
Note:
This engine is focused for easy to use API, currently some implementation still not efficient because Golang doesn't support:
- Class/Object Oriented Programming (solution: functional programming and `map[string] func()` will be used)
- Getter/setter (solution: use function as a getter/setter)
- Optional parameter (this will impact the performance of current engine implementation)

`interface{}` or `any` still being used massively to make Blackprint works without too much complexity in Golang, so please don't expect too high. Currently looking for contributor who can tidy up and make the implementation more efficient if possible.

Some implementation also use `reflect` for obtaining and setting value by using dynamic field name. It may be improved on the future, current implementation is focused to be similar with PHP and JS.

## License
MIT