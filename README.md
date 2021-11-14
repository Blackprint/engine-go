<p align="center"><a href="#" target="_blank" rel="noopener noreferrer"><img width="150" src="https://user-images.githubusercontent.com/11073373/141421213-5decd773-a870-4324-8324-e175e83b0f55.png" alt="Blackprint"></a></p>

<h1 align="center">Blackprint Engine for Golang</h1>
<p align="center">Run exported Blackprint on Golang environment.</p>

<p align="center">
    <a href='https://github.com/Blackprint/Blackprint/blob/master/LICENSE'><img src='https://img.shields.io/badge/License-MIT-brightgreen.svg' height='20'></a>
</p>

## Blackprint Engine for Golang
Note:
This engine is focused for easy to use API, currently some implementation still not efficient because Golang doesn't support:
- Getter/setter (solution: use function as a getter/setter)
- Optional parameter (solution: use `...interface{}`)

`interface{}` or `any` still being used massively to make Blackprint works without too much complexity in Golang, so please don't expect too high for the performance. Currently looking for contributor who can tidy up and make the implementation more efficient if possible, and without adding more complexity.

Some implementation also use `reflect` for obtaining and setting value by using dynamic field name. It may be improved on the future, current implementation is focused to be similar with PHP and JS.

## Documentation
> Warning: This project haven't reach it stable version (semantic versioning at v1.0.0)<br>
> But please try to use it and help the author for improving this project

> Need help writing '-'

## License
MIT