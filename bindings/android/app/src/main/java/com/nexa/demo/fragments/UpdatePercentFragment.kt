package com.nexa.demo.fragments

import androidx.annotation.LayoutRes
import androidx.fragment.app.Fragment

abstract class UpdatePercentFragment(@LayoutRes contentLayoutId: Int) : Fragment(contentLayoutId) {
    abstract fun updatePercent(allPercent: ArrayList<Int>)
    abstract fun updateImages(allImages: ArrayList<String>)
}