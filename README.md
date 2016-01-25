# WolfenGo

This is a clone of Wolfenstein3D written in Go, based on [Wolfenstein3DClone](https://github.com/BennyQBD/Wolfenstein3DClone) and licensed under [GNU/GPLv2](./LICENSE).

Pull requests are welcome.

## Plan
- [x] initial conversion
- [ ] fix remaining bugs
- [ ] add audio effects

# Building

```
git submodule init
git submodule update
make
```

Then you can run:
```
bin/wolfengo
```

There are some constants in `main.go` that can be toggled to enable further debugging/experimentation.

# Controls

Use `W`,`A`,`S`,`D` to move the player around and `E` to open doors; by clicking in the game window you will enable free mouse look, that can be disabled with `ESC`.

# History

Aside from some dead/unused code that I have dropped and bugs inadvertently introduced in the porting process, this is my ([gdm85](https://github.com/gdm85)) literal conversion of the [Java Wolfenstein3D clone by BennyQBD](https://github.com/BennyQBD/Wolfenstein3DClone); feel free to spin up the Java original version to check how identical and indistinguishable the two are.

Notable differences:
* map format has been changed, see relative section
* enabled VSync
* player does not seem able to shoot monsters, couldn't figure out where's the bug
* although extra shaders are included, they are not used by default in any way

Although [mathgl](https://github.com/go-gl/mathgl) could have been used for the 3D math/raycast operations, I preferred to keep the simpler original structures.

Some lessons learnt during the porting process:
* after 10 years I last experimented with OpenGL debugging, it's still hard. [apitrace](https://github.com/apitrace/apitrace) helped a lot
* Go's `int` is 64bit on 64bit platforms, a visual inspection doesn't tip off this easily and one could skip the fact that OpenGL needs to know the proper size (4 or 8 bytes)
* a minus sign here and there can screw up **a lot** of projections/translations

Development started on 4 January and completed with the first release on Github on 24th January with squashed/cleaned up commits.
See also [my blog post about WolfenGo initial release](https://medium.com/where-do-we-go-now/wolfengo-a-wolfenstein-3d-clone-in-go-6872af12469d) for extras about the development/porting process.

# Map format

The map format has been changed as well to be easier to edit via text editors (although the new format it's far from being final).

Some extensions have been added for other items (FPS lore quiz: where have you seen this map format already?)

The first lines of the map define walls:
```
wall1   {1.00,0.75,0.75,1.00}
wall2   {0.25,0.00,0.00,0.25}
```

The values in curly braces are texture coordinates from the tileset [WolfCollection.png](./res/textures/WolfCollection.png).
After that follows the `lengthmap` definition to indicate map size:
```
lengthmap       032
```
The value `32` means that this is a 32x32 map.

Subsequently there are the `MAP`, `PLANES` and `SPECIALS` sections, each of them describing with `n = 32` lines of `n = 32` characters either a wall value,
a floor/ceiling value or a special item. Wall characters start at `1` and (unfortunately) right now can as well go above `9`.

The special items that are currently supported are:
* `m` to indicate a small medkit
* `e` to indicate an enemy
* `d` to indicate a door
* `A` to indicate player start position
* `X` to indicate level exit

# Thanks

Obviously thanks to BennyQBD for the initial clone Java sources and also to https://github.com/go-gl/gl which - although not easy to master - is indeed in a good status for usage in Go OpenGL projects.
