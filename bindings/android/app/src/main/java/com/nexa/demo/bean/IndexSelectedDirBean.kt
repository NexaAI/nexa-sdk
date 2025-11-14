package com.nexa.demo.bean

import java.io.File

data class IndexSelectedDirBean(
    val dir: File,
    var isSelected: Boolean = false,
    var subImageFiles: ArrayList<String>? = null,
    var subVideoFiles: ArrayList<String>? = null
)