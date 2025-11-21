package com.nexa.demo.adapter

import androidx.fragment.app.Fragment
import androidx.fragment.app.FragmentManager
import androidx.fragment.app.FragmentPagerAdapter
import com.nexa.demo.fragments.UpdatePercentFragment

class IndexViewPagerAdapter(
    fm: FragmentManager,
    private val fragments: MutableList<UpdatePercentFragment>,
    private val titles: MutableList<String>
) : FragmentPagerAdapter(fm) {

    override fun getItem(position: Int): Fragment {
        return fragments.get(position)
    }

    override fun getCount(): Int {
        return fragments.size
    }

    override fun getPageTitle(position: Int): CharSequence? {
        return titles[position]
    }

    fun updatePercent(position: Int, allPercent: ArrayList<Int>) {
        fragments[position].updatePercent(allPercent)
    }
    fun updateImages(position: Int, allImages: ArrayList<String>) {
        fragments[position].updateImages(allImages)
    }
}